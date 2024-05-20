const
  resolv_conf_path* = "/etc/resolv.conf"
  resolv_conf_backup* = "/etc/anonsurf/resolv.conf.bak"
  dhclient_binary* = "/usr/sbin/dhclient"

  # https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/configuring_and_managing_networking/manually-configuring-the-etc-resolv-conf-file_configuring-and-managing-networking
  nm_hook_script* = "/etc/NetworkManager/conf.d/90-dns-none.conf"
  nm_hook_data* = "[main]\ndns=none\n"

type
  AddrFromParams* = object
    has_dhcp_flag*: bool
    list_addr*: seq[string]
