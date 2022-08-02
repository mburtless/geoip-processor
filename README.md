# geoip-processor
External Processor for Envoy that Adds GeoIP Data to HTTP Traffic

## Testing
```shell
docker-compose up -d
# test the external processor directly
grpcurl -plaintext -d '{"async_mode": false, "request_headers": {"headers": {"headers": [{"key": "x-forwarded-for", "value": "8.8.8.8"}]}}}' 127.0.0.1:8000 envoy.service.ext_proc.v3.ExternalProcessor/Process
# test via httpbin. Spoof IP by manually setting XFF header and note x-country-code in response
curl -H "X-Forwarded-For: 8.8.8.8" 127.0.0.1:10000/headers
```