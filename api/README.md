# API Proto files

## Add a proto template

```bash
kratos proto add api/tenants/v1/members.proto
```

## Generate the proto code

```bash
kratos proto client api/tenants/v1/members.proto
kratos proto client api/tenants/v1/invites.proto
```

## Generate the source code of service by proto file

```bash
kratos proto server api/tenants/v1/members.proto -t internal/service
```
