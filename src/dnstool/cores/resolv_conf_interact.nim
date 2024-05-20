import os
import dnstool_const
import utils
import network_manager_interact
import ../ cli / print


proc resolv_create_dhcp_only*() =
  # FIXME: if not remove resolv.conf, all addrs are kept, meanwhile remove it cause missing file
  # if not try_remove_file(resolv_conf_path):
  #   print_error("Failed to remove current DNS settings")
  #   return
  nm_remove_hook_script()
  # nm_reload_service()


proc resolv_create_addrs*(list_addr: seq[string]) =
  if len(list_addr) == 0:
    print_error("No valid addresses is provided")
    return

  if not tryRemoveFile(resolv_conf_path):
    print_error("Failed to remove " & resolv_conf_path & " to create new one.")
    return

  write_dns_to_system(list_addr)
  nm_create_hook_script()


proc resov_make_backup*() =
  # TODO review code logic here, move all code that handles resolv.conf to resolv_conf_interact
  if not fileExists(resolv_conf_path):
    print_error("Skip creating backup file. Missing " & resolv_conf_path)
    return

  if system_has_only_localhost():
    print_error("Skip creating backup file. System has only localhost.")
    return

  if fileExists(resolv_conf_backup):
    if not tryRemoveFile(resolv_conf_backup):
      print_error("Failed to remove " & resolv_conf_backup & " to create new one.")
      return

  if system_dns_file_is_symlink():
    #[
      /etc/resolv.conf is a symlink of resolvconf which is at
      /run/resolvconf/resolv.conf so we don't create backup here
      When there's no backup, dnstool will try create DHCP's DNS
      which should create a symlink if resolvconf is installed
    ]#
    print_error("Skip creating backup file because " & resolv_conf_path & " is a symlink.")
    return

  try:
    copyFile(resolv_conf_path, resolv_conf_backup)
  except:
    print_error("Failed to create backup file")
