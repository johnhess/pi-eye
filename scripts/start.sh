# This script starts the sniffer and webserver.
# On a Raspberry Pi, you may need to take additional steps.  
# See rpi_init.sh

# get the latest version and install it
git pull;
source setup_env.sh
go install pi-eye/hello

# start the sniffer as a daemon
echo "starting sniffer";
tshark -I -i mon0 -o wlan.enable_decryption:TRUE -o nameres.network_name:TRUE -o nameres.use_external_name_resolver:TRUE -o nameres.dns_pkt_addr_resolution:TRUE -T ek -j 'tcp dns ip' | sed '/^s*$/d' | hello &

# start the webserver as a daemon, too
cd visualization
echo "" > hist.json
python -m SimpleHTTPServer &>/dev/null &

chromium-browser --start-fullscreen "http://localhost:8000/" & 