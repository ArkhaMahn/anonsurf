const
  resolv_conf_path* = "/etc/resolv.conf"
  resolv_conf_backup* = "/etc/anonsurf/resolv.conf.bak"

  # https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/configuring_and_managing_networking/manually-configuring-the-etc-resolv-conf-file_configuring-and-managing-networking
  nm_hook_script* = "/etc/NetworkManager/conf.d/90-dns-none.conf"
  # https://people.freedesktop.org/~lkundrak/nm-docs/NetworkManager.conf.html
  nm_hook_dont_manage_resolvconf* = "[main]\ndns=none\nrc-manager=unmanaged\n"
  nm_hook_handle_resolvconf* = "[main]\ndns=default\nrc-manager=symlink\n"

type
  AddrFromParams* = object
    has_dhcp_flag*: bool
    list_addr*: seq[string]
