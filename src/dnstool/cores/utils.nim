import os
import dnstool_const
import strutils
import .. / cli / print
import sequtils
import net


proc convert_seq_addr_to_string(addresses: seq[string]): string =
  var
    dns_addr_in_text = ""
  for address in addresses:
    dns_addr_in_text &= "nameserver " & address & "\n"

  return dns_addr_in_text


proc has_only_localhost*(list_addr: seq[string]): bool =
  if len(list_addr) == 1 and (list_addr[0] == "localhost" or list_addr[0] == "127.0.0.1"):
    return true
  return false


proc system_dns_file_is_symlink*(): bool =
  if getFileInfo(resolv_conf_path, followSymlink = false).kind == pcLinkToFile:
    return true
  return false


proc parse_dns_addresses*(): seq[string] =
  for line in lines(resolv_conf_path):
    if line.startsWith("nameserver"):
      result.add(line.split(" ")[1])


proc system_has_only_localhost*(): bool =
  let
    addresses = parse_dns_addresses()

  return has_only_localhost(addresses)


proc write_dns_to_system*(list_dns_addr: seq[string]) =
  let
    deduplicated_list_addr = deduplicate(list_dns_addr)
  if len(deduplicated_list_addr) == 0:
    print_error("No valid DNS address is found.")
    return

  let
    dns_addr = convert_seq_addr_to_string(deduplicated_list_addr)
  try:
    writeFile(resolv_conf_path, "# Written by DNSTool\n" & dns_addr)
  except:
    print_error("Error while writing DNS address to " & resolv_conf_path)


proc get_list_valid_addresses*(list_addr: seq[string]): seq[string] =
  for each_addr in list_addr:
    if isIpAddress(each_addr):
      result.add(each_addr)

  return result


proc parse_addr_from_params*(params: seq[string]): AddrFromParams =
  var
    param_result = AddrFromParams(
      has_dhcp_flag: false
    )

  for value in params:
    if value == "dhcp":
      param_result.has_dhcp_flag = true
    elif isIpAddress(value):
      param_result.list_addr.add(value)

  return param_result
