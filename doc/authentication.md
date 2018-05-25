# Authentication

API calls can be authorized either by providing an **API Key** or an **OAuth Access Token**.

- Usage with HTTP `Authorization` Header: `Bearer XXXXX`
- Usage with gRPC [call credentials](https://grpc.io/docs/guides/auth.html#authentication-api) (in the `authorization` header): `Bearer XXXXX`
- Usage with MQTT: Username: Gateway ID or Application ID, Password: `XXXXX`

Here, `XXXXX` is either a valid **API Key** or a valid **OAuth Access Token**.

## API Keys

API keys are the simplest way of authorization. API keys do not expire, are revokable, and they are scoped to the entity they were generated from:

- User API keys
- Application API keys
- Gateway API keys
- Organization API keys

## OAuth Access Tokens

The Things Network uses the [OAuth 2.0 protocol](https://oauth.net/) for authentication and authorization. 

To use this method, you first need an **OAuth Client Registration**:

- The **Client ID** uniquely identifies the OAuth Client. Its restrictions are the same as for any other ID in TTN.
- The **Description** is shown to the user when you request authorization.
- The **Scope** indicates what actions your OAuth Client is allowed to perform. This is shown to the user when you request authorization. You can select the actions your OAuth Client needs on registration, a full list can also be found in our [source code](https://github.com/TheThingsIndustries/ttn/blob/ttn-master/api/rights.proto).
- The **Redirect URI** is where the user is redirected after authorizing your OAuth Client.
- The **Client Secret** is issued when your OAuth Client Registration is accepted by a network admin.

After your OAuth Client Registration is accepted, you can **request authorization** by sending the user to the **Authorization URL**:

```
https://<HOSTNAME>/oauth/authorize?client_id=<CLIENT-ID>&redirect_uri=<REDIRECT-URI>&state=<STATE>&response_type=code
```

- The `HOSTNAME` is the hostname of the Identity Server.
- The `client_id` is the **Client ID** of your OAuth Client.
- The `response_type` is always `code`.
- The `redirect_uri` must exactly match the **Redirect URI** of your OAuth Client Registration if supplied.
  - We allow multiple **Redirect URIs** in your OAuth Client Registration in the future, in which case the `REDIRECT-URI` must exactly match one of those.
- The optional `scope` is ignored by the Identity Server. All scopes defined in your OAuth Client Registration will be requested.
- The optional `state` can be used to mitigate CSRF attacks. It is recommended to supply this.

The Identity Server will prompt the user with a view asking to authorize your OAuth Client. They will see the **Client ID**, **Description**, requested **Scope** and **Redirect URI**. If they accept the authorization, they will be redirected to your **Redirect URI** with an **Authorization Code**:

```
https://<REDIRECT-URI>/?code=<AUTHORIZATION-CODE>
```

Your OAuth Client can exchange this **Authorization Code** for an **OAuth Access Token** by making a `POST` request to the **Token URL**:

```
https://<HOSTNAME>/oauth/token
```

The request must use **Basic Auth** ([RFC7617](https://tools.ietf.org/html/rfc7617)) with the **Client ID** as username and the **Client Secret** as password.

The **Authorization Code** is sent in the request payload:

```
{
	"code": "<AUTHORIZATION-CODE>", 
	"grant_type": "authorization_code"
}
```

The response contains the **OAuth Access Token** and an indication of when it expires. If the network admin gave your OAuth Client the **Refresh Token** grant, the response also contains a **Refresh Token**.

```
{
	"access_token": "XXXXX", 
	"token_type": "bearer", 
	"expires_in": "3600",
	"refresh_token": "YYYYY"
}
```

You can now use the **OAuth Access Token** until it expires. 

If you have a **Refresh Token**, you can exchange this for a new **OAuth Access Token** after the old one expires by making another `POST` request to the **Token URL**, similar to the exchange of the **Authorization Code** you did before:

```
https://<HOSTNAME>/oauth/token
```

The request must use **Basic Auth** ([RFC7617](https://tools.ietf.org/html/rfc7617)) with the **Client ID** as username and the **Client Secret** as password.

The **Refresh Token** is sent in the request payload:

```
{
	"code": "<REFRESH-TOKEN>", 
	"grant_type": "refresh_token"
}
```

The response again contains the **OAuth Access Token** and an indication of when it expires. The response also contains a new **Refresh Token**.

```
{
	"access_token": "XXXXX", 
	"token_type": "bearer", 
	"expires_in": "3600",
	"refresh_token": "YYYYY"
}
```
