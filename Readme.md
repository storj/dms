# DMS - Dead (Wo)man Switch
Simple aggregate dead (wo)man switch

### Use

#### Register
```bash
curl --user <basic_auth_user>:<basic_auc_password> -d "environment=<environment_name>" <dms_endpoint>/register
```

#### Ingest
```bash
curl -H "Authorization: Bearer <jwt_token>" -XPOST <dms_endpoint>/ingest
```

#### Pingdom
```bash
curl <dms_endpoint>/pingdom
```

#### Status
```bash
curl --user <basic_auth_user>:<basic_auc_password> <dms_endpoint>/status
```

#### Incidents
```bash
curl --user <basic_auth_user>:<basic_auc_password> <dms_endpoint>/incidents
```

### Test
```bash
go test -parallel 1 -v ./pkg/routes/... ./pkg/storage/... -cover
```

### Author
- [kristaxox](https://github.com/kristaxox)
