package bind

const (
	supernodeInclude   = "include \"/etc/bind/named.supernode.conf\";"
	namedSupernodeConf = `acl "supernodes" {
${OTHER_SUPERNODE_IPS}
};

masters "supernodes" {
${OTHER_SUPERNODES_IPS}
};

zone "${SUPERNODE_ZONE}.mesh" {
    type master;
    also-notify { supernodes; };
    allow-transfer { supernodes; };
    file "/etc/bind/${SUPERNODE_ZONE}.mesh.zone";
};

// zone "othersupernode.mesh" {
//     type slave;
//     masters { 1.1.1.1; };
//     allow-notify { 1.1.1.1; };
//     masterfile-format text;
//     file "/etc/bind/othersupernode.mesh.zone";
// };
`

	supernodeSlaveZone = ``

	supernodeMasterZone = `$TTL 60
$ORIGIN ${SUPERNODE_ZONE}.mesh.
@  SOA  ns0.${SUPERNODE_ZONE}.mesh. root.${SUPERNODE_ZONE}.mesh. (
  ${SERIAL} ; Serial
  3600      ; Refresh
  300       ; Retry
  604800    ; Expire
  60 )      ; TTL
;
@           NS ns0
ns0        A  ${NODE_IP}
${NODE_NAME} A  ${NODE_IP}
`

	localMeshZone = `$TTL 60
$ORIGIN local.mesh.
@  SOA ns.local.mesh. master.local.mesh. (
  ${SERIAL}   ; Serial
  3600        ; Refresh
  300         ; Retry
  604800      ; Expire
  60 )        ; TTL
)
@     NS ns0
ns0 A  ${NODE_IP}
${EXTRA_HOSTS}
`

	meshZone = `$TTL 60
$ORIGIN mesh.
@  SOA ns.mesh. master.mesh. (
  ${SERIAL}   ; Serial
  3600        ; Refresh
  300         ; Retry
  604800      ; Expire
  60 )        ; TTL

NS ns0.local
ns0.local A  ${NODE_IP}
local     NS ns0.local
`

	meshZoneSupernode = `         NS ns0.${SUPERNODE_MESH_NAME}
ns0.${SUPERNODE_MESH_NAME} A  ${SUPERNODE_MESH_IP}
${SUPERNODE_MESH_NAME}      NS ns0.${SUPERNODE_MESH_NAME}
`

// Always:
// local.mesh.zone will be generated from the template
// mesh.zone will be generated from the template
// named.conf will stay as is

// If supernode is enabled:
// We might add the include statement if supernode is enabled
// named.supernode.conf will be generated from the template
// {supernodename}.mesh.zone will be generated from the template
// {othersupernodename}.mesh.zone slave will be generated from the template? /etc/bind/othersupernode.mesh.zone???
