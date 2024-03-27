go build
mkdir -p ~/.local/share/tezaurs ~/.local/bin 
mv ./tezaurs ~/.local/bin/
cp ./wordlist.txt ~/.local/share/tezaurs/
echo "Tēzaurs ieinstalēts!"
