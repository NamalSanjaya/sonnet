### Mserver API input data formats examples
1. | /ms/set-ds1/:userId | PUT |
   | ------------------- | --- |  

```json
 {
	"username":"namal-sanjaya-12",
	"email": "namaltest@mail.com",
	"blockuserlist": ["samal44", "yalu88", "kamal"],
	"histtbs": { 
      "usr-11": { "tx2rx": "my-link-u11", "rx2tx":"u11-link-my"}, 
      "usr-23": { "tx2rx": "my-link-u23", "rx2tx":"u23-link-my"},
      "usr-19": { "tx2rx": "my-link-u19", "rx2tx":"u19-link-my"}
	    }
}
```
2. | /ms/set-blockuser/:userId?userid=someid | PUT |
   | --------------------------------------- | --- |
 
3. | /ms/set-newcontact-ds1/:userId?userid=newUserId | PUT |
   | ----------------------------------------------- | --- |
   
```json
{ 
     "tx2rx": "tx-link-rx", 
     "rx2tx":"rx-link-tx"
}
```
4. | /ms/del-blockuser/:userId?userid=rmSomeid | PUT |
   | ----------------------------------------- | --- |
   
5. | /ms/set-newcontact-ds2/:userId?userid=someToUserId | PUT |
   | -------------------------------------------------- | --- |
   ```json
   { 
	 "tx2rx": "5173cb67-652b-46d6-8b4b-342a0eba1cdc", 
	 "rx2tx":"44aed4af-b121-468b-8ac8-499b36a63aa2"
   }
   ```
 6. | /ms/set-lastread/:userId?tohist=someToUserId&nxtread=107 | PUT |
    | -------------------------------------------------------- | --- |
    
    
 7. | /ms/set-lastmsg/:userId?hist=myHistTb&latestmsg=1137 | PUT |
    | ---------------------------------------------------- | --- |
    
    
    
