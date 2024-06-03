import httpclient
import xmltree
import htmlparser
import strutils
import print_utils


proc parseIPFromTorServer*(data: string, ip_addr, status: var string) =
  #[
    Get IP address and current status from Tor server
  ]#

  ip_addr = parseHtml(data).findAll("strong")[0].innerText
  status = parseHtml(data).findAll("title")[0].innerText

#[
  Check current public IP using https://check.torproject.org/ 
]#
proc check_status_from_tor_server(ip_addr, status: var string): bool =
  const
    target = "https://check.torproject.org/"
  var
    client = newHttpClient()

  try:
    let resp = client.get(target)
    parseIPFromTorServer(resp.body, ip_addr, status)
    return true
  except Exception:
    cli_send_msg(Err, "Check IP error", getCurrentExceptionMsg())
    return false
  finally:
    client.close()

#[
  Check IP via tor
]#
proc cli_check_current_ip*() =
  var
    ip_addr, status: string

  if check_status_from_tor_server(ip_addr, status):
    if "Sorry" in status:
      cli_send_msg(Warn, "Your public IP is " & ip_addr, "You are not under Tor network ")
    else:
      cli_send_msg(Ok, "Your public IP is " & ip_addr, "You are under Tor network")
