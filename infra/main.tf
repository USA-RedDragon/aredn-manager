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
      version = "4.67.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.12.0"
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

# Create an IAM policy allowing EC2 to create Cloudwatch log entries
resource "aws_iam_policy" "ec2-cloudwatch-logs" {
  name        = "${var.server-name}-ec2-cloudwatch-logs"
  description = "Allow EC2 to create Cloudwatch log entries"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": [
                "arn:aws:logs:${var.region}:*:log-group:${aws_cloudwatch_log_group.log-group.name}",
                "arn:aws:logs:${var.region}:*:log-group:${aws_cloudwatch_log_group.log-group.name}:log-stream:*"
            ]
        }
    ]
}
EOF
}

# Create an IAM role for EC2 to assume
resource "aws_iam_role" "ec2-cloudwatch-logs" {
  name               = "${var.server-name}-ec2-cloudwatch-logs"
  assume_role_policy = data.aws_iam_policy_document.ec2-cloudwatch-logs.json
}

# Attach the policy to the role
resource "aws_iam_role_policy_attachment" "ec2-cloudwatch-logs" {
  role       = aws_iam_role.ec2-cloudwatch-logs.name
  policy_arn = aws_iam_policy.ec2-cloudwatch-logs.arn
}

# Create a policy document allowing EC2 to assume the role
data "aws_iam_policy_document" "ec2-cloudwatch-logs" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

# Create an instance profile for EC2 to assume
resource "aws_iam_instance_profile" "ec2-cloudwatch-logs" {
  name = "${var.server-name}-ec2-cloudwatch-logs"
  role = aws_iam_role.ec2-cloudwatch-logs.name
}

resource "aws_cloudwatch_log_group" "log-group" {
  name = var.server-name
  # Retention in days
  retention_in_days = 30
}

# Add a persistent ebs gp2 volume of 8GB
resource "aws_ebs_volume" "ebs" {
  size              = 8
  type              = "gp2"
  availability_zone = "${var.region}a"
  encrypted         = true
  final_snapshot    = true
}

resource "aws_volume_attachment" "ebs_att" {
  device_name                    = "/dev/sdf"
  volume_id                      = aws_ebs_volume.ebs.id
  instance_id                    = aws_instance.node.id
  stop_instance_before_detaching = true
}

resource "aws_instance" "node" {
  ami           = data.aws_ami.ubuntu-jammy.id
  instance_type = var.instance-type

  user_data = templatefile("${path.module}/user-data.sh", {
    server_name                 = var.server-name
    server_lon                  = var.server-lon
    server_lat                  = var.server-lat
    server_gridsquare           = var.server-gridsquare
    map_config_json             = var.map-config-json
    wireguard_tap_address       = var.wireguard_tap_address
    region                      = var.region
    awslogs-group               = aws_cloudwatch_log_group.log-group.name
    wireguard_peer_publickey    = var.wireguard_peer_publickey
    wireguard_server_privatekey = var.wireguard_server_privatekey
    node_ip                     = var.node_ip
    supernode_zone              = var.supernode_zone
    pg_host                     = var.pg_host
    pg_user                     = var.pg_user
    pg_pass                     = var.pg_password
    pg_db                       = var.pg_db
    session_secret              = var.session_secret
    password_salt               = var.password_salt
    extra_cors_hosts            = var.extra_cors_hosts
    init_admin_user_password    = var.init_admin_user_password
    supernode                   = var.supernode
  })
  user_data_replace_on_change = true

  vpc_security_group_ids = [aws_security_group.allow-vpn.id]

  iam_instance_profile = aws_iam_instance_profile.ec2-cloudwatch-logs.name

  key_name = aws_key_pair.key.key_name

  availability_zone = "${var.region}a"

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
    from_port   = 51820
    to_port     = 51820
    protocol    = "udp"
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
