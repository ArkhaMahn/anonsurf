import os
import dnstool_const
import utils
import network_manager_interact
import ../ cli / print


proc resolv_create_dhcp_only*() =
  if not try_remove_file(resolv_conf_path):
    print_error("Failed to remove current DNS settings")
    return
  nm_create_hook_dhcp()


proc resolv_create_addrs*(list_addr: seq[string]) =
  if len(list_addr) == 0:
    print_error("No valid addresses is provided")
    return

  if not tryRemoveFile(resolv_conf_path):
    print_error("Failed to remove " & resolv_conf_path & " to create new one.")
    return

  write_dns_to_system(list_addr)
  nm_create_hook_unmanage()


proc resov_restore_from_backup*() =
  #[
    Remove /etc/resolv.conf
    Move backup file to /etc/resolv.conf
  ]#
  if not tryRemoveFile(resolv_conf_path):
    print_error("Restore backup: Failed to remove " & resolv_conf_path & " to restore backup.")
    # It might cause error depends on file type. Return instead?

  if fileExists(resolv_conf_backup):
    try:
      moveFile(resolv_conf_backup, resolv_conf_path)
    except:
      print_error("Restore backup: Failed to restore backup")


proc resov_make_backup*() =
  if not fileExists(resolv_conf_path):
    print_error("Create backup: Skip creating backup file. Missing " & resolv_conf_path)
    return

  if system_has_only_localhost():
    print_error("Create backup: Skip creating backup file. Only found localhost nameserver.")
    return

  if fileExists(resolv_conf_backup):
    if not tryRemoveFile(resolv_conf_backup):
      print_error("Create backup: Failed to remove " & resolv_conf_backup & " to create new one.")
      return

  if system_dns_file_is_symlink():
    #[
      /etc/resolv.conf is a symlink of resolvconf which is at
      /run/resolvconf/resolv.conf so we don't create backup here
      When there's no backup, dnstool will try create DHCP's DNS
      which should create a symlink if resolvconf is installed
    ]#
    print_error("Create backup: Skip creating backup file because " & resolv_conf_path & " is a symlink.")
    return

  try:
    copyFile(resolv_conf_path, resolv_conf_backup)
  except:
    print_error("Create backup: Failed to create backup file")
