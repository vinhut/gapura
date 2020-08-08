#/bin/bash

while IFS=, read -r email pass;
do 
	curl -XPOST -d "email=$email&password=$pass" $1/user
done < username.txt 
