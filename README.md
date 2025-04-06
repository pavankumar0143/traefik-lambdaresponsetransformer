# traefik-lambdaresponsetransformer

This middleware transforms a JSON-formatted Lambda proxy response into a standard HTTP response.
  It handles decoding base64 bodies, applying headers, and setting appropriate CORS headers based on the original request.
  Only supports lambda running in local environment. For local development, this plugin is required to be used in conjunction with lambdarequesttransformer, lambdaauthorizer plugins.