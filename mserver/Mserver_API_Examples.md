### Mserver API input data formats examples
1. | /ms/set-ds1/:userId | PUT |
   | ------------------- | --- |  
Request body
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
   
Request body
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
   
   Request body
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
     
 8. | /ms/del-msg/:userId?hist=myHistTb&delmsg=4351 | PUT |
    | --------------------------------------------- | --- |
    
 9. | /ms/load-msg?userid=myId&touserid=friendId&hist=myHistTb&tohist=toHistTb&start=1025&end=1457 | GET |
    | -------------------------------------------------------------------------------------------- | --- |
    
Response Body
```json
  {
     "Err": 0,
      "msgs": [
         {
           "Timestamp": 451,
            "Data": "namal, now in 451",
            "Size": 10
         },
         {
            "Timestamp": 438,
            "Data": "namal, hello bye 438",
            "Size": 40
         }
       ]
  }
 ```
#### 10 
  | /ms/load-all-contactnmsgs?userid=myId | GET |
  | ------------------------------------- | --- |

Response Body

```json
{  
  "Err": 0
  "FriendUserId-1" : {
  		    "tx2rx" :  {
		    		 "histId"   : "5173cb67-652b-46d6-8b4b-342a0eba1cdc",
				 "LastMsg"  : 12405
 				 "Size"     : 2048
				}
  		    "rx2tx" : { 
		    		"histId"    : "44aed4af-b121-468b-8ac8-499b36a63aa2",
				"LastMsg"   : 7009
				"LastRead"  : 6990
				"Size"      : 1900
			      }
		    "userId": "FriendUserId-1",
		    "username": "my_friend_username",
		    "email" : "friend@mail.com",
		    "imgId": "someImgId",
		    "lastest_updated" : "${max(lastmsg_o, lastmsg_f)}"
		    "loaded_upto" : "min_timestamp_loaded_send_to_frontend"
		    "content": [
		    		 {"Timestamp": 5793, "data": "-----msg-content-----", "Size": 38, "link": "f"},
		    		 {"Timestamp": 5797, "data": "-----msg-content-----", "Size": 45, "link": "f"},
		    		 {"Timestamp": 6041, "data": "-----msg-content-----", "Size": 18, "link": "o"},
		    	       ]
		  }

  "FriendUserId-7" : {}
  "FrndUserId-3"   : {}
}

```
**note**
* order of friend's Ids(most recent ids to the top) important as it imples the which user send msg recently.(we can use `lastest_updated` field to maintain this behaviour
* lastest_updated - used to sort the friend's chat list. with whom owner recently chat with. calculte from LastMsg field.
* LastRead - when last read did by the owner.
* `content.link` use to identify whose data block that is.("o" - owner , "f" - friend)
* `content` should come in order from backend. latest msgs to the top.(for a friend)
* `loaded_upto` min timestamp loaded to UI. we need to fetch from here onward from backend when it need more past msgs.
