#/bin/bash

while IFS=, read -r email pass;
do 
	username=`echo $email | cut -d@ -f1`
	curl -XPOST -d "email=$email&password=$pass&username=$username" $1/user?service=internal
done < username.txt 

kubectl -n auth-test exec -it mongodb-auth-0 -- mongo --quiet -umongoadmin -psecret --authenticationDatabase "admin" --eval "db.users.createIndex({_id:1,username:1,email:1})" pamulangdb
