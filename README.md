# Remote Agile Toolbox - Auth Service

Authorization service using social sign in via GitHub, etc. for the remote-agile-toolbox application.


### auth flow outline

1. goto login url 
2. get redirected to selected oauth provider, specified in initial login url
3. login with user credentials at auth provider
4. oauth provider redirects to callback handler with `code` (response is secured with `state`)
5. callback handler exchanges received `code` for provider-access-token
6. get user info from provider using exchanged provider-access-token
7. redirect to client application
8. client application handles received `token-id` and `exchange-code` for `app-access-token`