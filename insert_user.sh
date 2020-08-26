#/bin/bash

while IFS=, read -r email pass;
do 
	username=`echo $email | cut -d@ -f1`
	curl -XPOST -d "email=$email&password=$pass&username=$username" $1/user
done < username.txt 
