# Meteor Client Connection

## Brief Introduction

    If you want to connected to a meteor client, there is something to follow.
    
    - Meteor client will first try to connected to `{host}/sockjs/info` to see if meteor support websocket or not.
    - If meteor server response contain `websocket: true`, then Meteor client will try to use websocket to connect to meteor server.
    - Else, meteor will use sockjs to connect to meteor server
    - The websocket path client will try to connected is in path: /sockjs/{random}/{random}/websocket
    
    So, server will need to expose the following route:
        - /websocket: For bot to connected
        - /sockjs/info: to return info to let client to connect to websocket
        - /sockjs/{random}/{random}/websocket: to connect websocket client.
        
## Client convention

    - Server need to send a `o` string to Meteor client to start a conversation
    - All server return Json need to prefix a `a` char
    

