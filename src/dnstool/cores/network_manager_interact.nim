import os
import osproc
import strutils
import dnstool_const
import .. / cli / print


proc nm_reload_service() =
  if execCmd("/usr/bin/systemctl reload NetworkManager") != 0:
    print_error("Failed to reload NetworkManager service")


proc nm_get_dns_servers*(): seq[string] =
  #[
    Get DNS addresses from DHCP server. This function parses output from command
      /usr/bin/nmcli dev show | grep DNS
    Output be like (space)
      IP4.DNS[1]:                             192.168.58.1,1.1.1.1
    https://infotechys.com/change-dns-settings-using-the-nmcli-utility/
  ]#
  return execProcess("/usr/bin/nmcli dev show | grep DNS").replace("\n", "").split(" ")[^1].split(",")


proc nm_create_hook_unmanage*() =
  #[
    Hookscript will prevent NetworkManager to overwrite configuration
    https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/configuring_and_managing_networking/manually-configuring-the-etc-resolv-conf-file_configuring-and-managing-networking
    Create new hook script that has exec in file's permissions, and then restart NetworkManager service
  ]#
  try:
    writeFile(nm_hook_script, nm_hook_dont_manage_resolvconf)
    setFilePermissions(nm_hook_script, {fpUserWrite, fpUserRead, fpGroupWrite, fpGroupRead, fpOthersRead})
    nm_reload_service()
  except:
    print_error("Failed to write hook script to " & nm_hook_script)


proc nm_create_hook_dhcp*() =
  #[
    Hookscript will prevent NetworkManager to overwrite configuration
    https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/configuring_and_managing_networking/manually-configuring-the-etc-resolv-conf-file_configuring-and-managing-networking
    Create new hook script that has exec in file's permissions, and then restart NetworkManager service
  ]#
  try:
    writeFile(nm_hook_script, nm_hook_handle_resolvconf)
    setFilePermissions(nm_hook_script, {fpUserWrite, fpUserRead, fpGroupWrite, fpGroupRead, fpOthersRead})
    nm_reload_service()
  except:
    print_error("Failed to write hook script to " & nm_hook_script)
