1. | /ms/del-msg/:userId?tohist=myHistTb&delmsg=delMsgIndex | PUT |
   | ------------------------------------------------------ | --- |
**logic**
```
if(delMsgIndex == lastmsg)
  if(lastread == lastmsg)
    lastread = Memory[-2]
   lastmsg = Memory[-2]
if(delMsgIndex < lastMsg)
  if(lastread == delMsgIndex)
    lastread = Memory[ adjecent index before delMsgIndex ]

if(delMsgIndex <= lastdeleted)
  return
else
  remove `delMsgIndex` from redis memory
```
