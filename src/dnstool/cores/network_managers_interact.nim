import os
import osproc
import strutils
import dnstool_const
import .. / cli / print


proc nm_get_dns_servers*(): seq[string] =
  #[
    Get DNS addresses from DHCP server. This function parses output from command
      /usr/bin/nmcli dev show | grep DNS
    Output be like (space)
      IP4.DNS[1]:                             192.168.58.1,1.1.1.1
    https://infotechys.com/change-dns-settings-using-the-nmcli-utility/
  ]#
  # FIXME handle seq contains only 1 empty addr like @[""]
  return execProcess("/usr/bin/nmcli dev show | grep DNS").replace("\n", "").split(" ")[^1].split(",")


proc nm_create_hook_script*() =
  #[
    Hookscript will prevent NetworkManager to overwrite configuration
    https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/configuring_and_managing_networking/manually-configuring-the-etc-resolv-conf-file_configuring-and-managing-networking
    Create new hook script that has exec in file's permissions, and then restart NetworkManager service
  ]#
  if not fileExists(nm_hook_script):
    try:
      writeFile(nm_hook_script, nm_hook_data)
      setFilePermissions(nm_hook_script, {fpUserExec, fpUserRead, fpGroupExec, fpGroupRead, fpOthersRead, fpOthersExec})
      discard execCmd("/usr/bin/systemctl reload NetworkManager")
    except:
      print_error("Failed to write hook script to " & nm_hook_script)
  else:
    print_error("Hook script " & nm_hook_script & " already existed")


proc nm_remove_hook_script*() =
  if not tryRemoveFile(nm_hook_script):
    print_error("Failed to remove hook script at " & nm_hook_script)
