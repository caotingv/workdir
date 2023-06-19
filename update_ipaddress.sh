#!/bin/bash

die() {
  local msg=$1
  local code=${2:-1}
  echo -e "\033[31m$msg \033[0m" >&2
  exit "$code"
}

parse_params() {
  # default values of variables set from params
  netcard=""
  ipaddr=""
  prefix=""
  gateway=""
  dns=""
  current_ipaddr=""
  global_vars_file="../etc_example/global_vars.yaml"
  
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -n|--netcard) netcard="${2-}" ; shift ;;
      -i|--ipaddr) ipaddr="${2-}" ; shift ;;
      -p|--prefix) prefix="${2-}" ; shift ;;
      -g|--gateway) gateway="${2-}" ; shift ;;
      -d|--dns) dns="${2-}" ; shift ;;
      -c|--current_ipaddr) current_ipaddr="${2-}" ; shift ;;
      -?*) die "Unknown option: $1" ;;
      *) break ;;
    esac
    shift
  done

  if [[ -z $netcard || -z $ipaddr || -z $prefix || -z $gateway ]]; then
    die "eg: ./$(basename "${BASH_SOURCE[0]}") --netcard eth_name --ipaddr 192.168.100.10 --prefix 24 --gateway 192.168.100.254"
  fi
}

gat_current_ipaddr() {
  if [[ $current_ipaddr == "" ]]; then
    current_ipaddr=$(grep -Eo 'IPADDR=[0-9.]+' /etc/sysconfig/network-scripts/ifcfg-$netcard | cut -d= -f2)
  fi

  if [[ -z $current_ipaddr || $current_ipaddr == 169.168* ]]; then
    die "Failed to get the current IP address, Set it by adding the --current_ipaddr parameter." 
  fi

  echo "The current IP address is: $current_ipaddr"
}

update_network() {
  sudo nmcli connection modify $netcard ipv4.addresses $ipaddr/$prefix
  sudo nmcli connection modify $netcard ipv4.gateway $gateway
  if [ ! -z "${dns-}" ]; then
    sudo nmcli connection modify $netcard ipv4.dns $dns
  fi
  sudo nmcli connection down $netcard >/dev/null && sudo nmcli connection up $netcard >/dev/null

  result=$(ip address show $netcard | grep "$ipaddr/$prefix" | wc -l)
  if [ $result -eq 0 ]; then
    die "netcard $netcard Change failed!!!" 
  fi
}

update_config_file() {
  # btserver
  if [[ -f /etc/klcloud/btserver/torrent.ini ]]; then
    if [[ $current_ipaddr != $ipaddr ]]; then
      sudo sed -i "s/${current_ipaddr}/${ipaddr}/g" /etc/klcloud/btserver/torrent.ini
      if ! grep -q "$ipaddr" /etc/klcloud/btserver/torrent.ini; then
        die "Update btserver config failed!"
      fi
      docker restart btserver btserver_tracker
    fi
  fi

  # fsd_console
  if [[ -f /etc/klcloud/btserver/torrent.ini ]]; then
    if [[ $current_ipaddr != $ipaddr ]]; then
      sudo sed -i "s/${current_ipaddr}/${ipaddr}/g" /etc/klcloud/fsd_console/fsd_console_start.sh
      result=$(cat /etc/klcloud/fsd_console/fsd_console_start.sh | grep $ipaddr | wc -l)
      if [ $result -eq 0 ]; then
        die "update fsd_console config failed!!!" 
      fi
      docker restart fsd_console_web
    fi
  fi

  # fsd_database
  java_db_name="klcloud_fsd_edu"
  mariadb_root_username=$(grep -E '^mariadb_root_username:' "$global_vars_file" | cut -d':' -f2 | awk '{print $1}')
  mariadb_root_password=$(grep -E '^mariadb_root_password:' "$global_vars_file" | cut -d':' -f2 | awk '{print $1}')
  select_setting_sql="select info from sys_setting where uuid = '5b2d43b2-5ff3-4731-a2e6-0854be87a8c7';"
  setting_info=$(docker exec -it mariadb mysql -N -s -u"$mariadb_root_username" -p"$mariadb_root_password" -e "use ${java_db_name}; $select_setting_sql")
  
  if [[ $setting_info == *"$current_ipaddr"* ]]; then
    new_setting_info=$(echo "$setting_info" | sed "s/${current_ipaddr}/${ipaddr}/g")
    update_setting_sql="UPDATE sys_setting SET info = '${new_setting_info}' WHERE uuid = '5b2d43b2-5ff3-4731-a2e6-0854be87a8c7';" 
    update_server_sql="UPDATE des_server SET ip = '${ipaddr}';"

    docker exec -it mariadb mysql -u"$mariadb_root_username" -p"$mariadb_root_password" -e "use ${java_db_name}; $update_setting_sql $update_server_sql"
  fi
}

parse_params "$@"

deploy_version=$(grep -E '^deploy_edu:' $global_vars_file | cut -d':' -f2)
if [ $deploy_version != true ]; then
  die "This script is only available in a single node Edition VOI environment!"
fi

gat_current_ipaddr
update_network
update_config_file
