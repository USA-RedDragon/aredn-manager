terraform {
  cloud {
    organization = "Personal-McSwain"

    workspaces {
      name = "aredn-cloud-node"
    }
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.49.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "3.31.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.4"
    }
  }
}

provider "aws" {
  region = var.region
}

data "aws_ami" "ubuntu-jammy" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_eip" "ip" {
  instance = aws_instance.node.id
  vpc      = true
}

resource "aws_instance" "node" {
  ami           = data.aws_ami.ubuntu-jammy.id
  instance_type = var.instance-type

  user_data = templatefile("${path.module}/user-data.sh", {
    server_name        = var.server-name
    configuration_json = var.configuration-json
  })

  vpc_security_group_ids = [aws_security_group.allow-vpn.id]

  key_name = aws_key_pair.key.key_name

  root_block_device {
    volume_type = "gp2"
    volume_size = var.disk-size
  }

  tags = {
    Name = var.server-name
  }
}

resource "tls_private_key" "key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "key" {
  key_name   = var.server-name
  public_key = tls_private_key.key.public_key_openssh
}


resource "aws_security_group" "allow-vpn" {
  name        = "${var.server-name}-vpn"
  description = "Security Group for VTun VPN"

  ingress {
    from_port   = 5525
    to_port     = 5525
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
