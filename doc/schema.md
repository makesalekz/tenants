## Group:

|    Field    |   Type    | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |          StructTag           | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id          | int       | false  | false    | false    | false   | false         | false     | json:"id,omitempty"          |          0 |         |
| created_at  | time.Time | false  | false    | false    | true    | false         | true      | json:"created_at,omitempty"  |          0 |         |
| updated_at  | time.Time | false  | false    | false    | true    | true          | false     | json:"updated_at,omitempty"  |          0 |         |
| deleted_at  | time.Time | false  | true     | true     | false   | false         | false     | json:"deleted_at,omitempty"  |          0 |         |
| identity_id | uuid.UUID | true   | false    | false    | false   | false         | true      | json:"identity_id,omitempty" |          0 |         |
| tenant_id   | int64     | false  | false    | false    | false   | false         | true      | json:"tenant_id,omitempty"   |          0 |         |
| name        | string    | false  | false    | false    | false   | false         | false     | json:"name,omitempty"        |          1 |         |
| description | string    | false  | false    | false    | false   | false         | false     | json:"description,omitempty" |          0 |         |


|  Edge   |  Type  | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| tenant  | Tenant | true    | groups  | M2O      | true   | false    |         |
| members | Member | false   |         | M2M      | false  | true     |         |

## Invite:

|    Field    |       Type        | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |          StructTag           | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id          | int               | false  | false    | false    | false   | false         | false     | json:"id,omitempty"          |          0 |         |
| tenant_id   | int64             | false  | false    | false    | false   | false         | true      | json:"tenant_id,omitempty"   |          0 |         |
| code        | uuid.UUID         | false  | false    | false    | false   | false         | true      | json:"code,omitempty"        |          0 |         |
| email       | string            | false  | false    | false    | false   | false         | true      | json:"email,omitempty"       |          0 |         |
| user_id     | int64             | false  | true     | true     | false   | false         | false     | json:"user_id,omitempty"     |          0 |         |
| status      | enum.InviteStatus | false  | false    | false    | true    | false         | false     | json:"status,omitempty"      |          0 |         |
| created_at  | time.Time         | false  | false    | false    | true    | false         | true      | json:"created_at,omitempty"  |          0 |         |
| updated_at  | time.Time         | false  | false    | false    | true    | false         | false     | json:"updated_at,omitempty"  |          0 |         |
| role_id     | int64             | false  | true     | false    | false   | false         | false     | json:"role_id,omitempty"     |          0 |         |
| resource    | string            | false  | true     | false    | false   | false         | false     | json:"resource,omitempty"    |          0 |         |
| resource_id | int64             | false  | true     | false    | false   | false         | false     | json:"resource_id,omitempty" |          0 |         |


|  Edge  |  Type  | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| tenant | Tenant | true    | invites | M2O      | true   | false    |         |

## Member:

|    Field    |   Type    | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |          StructTag           | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id          | int       | false  | false    | false    | false   | false         | false     | json:"id,omitempty"          |          0 |         |
| deleted_at  | time.Time | false  | true     | true     | false   | false         | false     | json:"deleted_at,omitempty"  |          0 |         |
| identity_id | uuid.UUID | true   | false    | false    | false   | false         | true      | json:"identity_id,omitempty" |          0 |         |
| tenant_id   | int64     | false  | false    | false    | false   | false         | true      | json:"tenant_id,omitempty"   |          0 |         |
| user_id     | int64     | false  | false    | false    | false   | false         | true      | json:"user_id,omitempty"     |          0 |         |
| created_at  | time.Time | false  | false    | false    | true    | false         | true      | json:"created_at,omitempty"  |          0 |         |


|  Edge  |  Type  | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| tenant | Tenant | true    | members | M2O      | true   | false    |         |
| groups | Group  | true    | members | M2M      | false  | true     |         |

## Tenant:

|   Field    |      Type       | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |          StructTag          | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id         | int64           | false  | false    | false    | false   | false         | true      | json:"id,omitempty"         |          0 |         |
| deleted_at | time.Time       | false  | true     | true     | false   | false         | false     | json:"deleted_at,omitempty" |          0 |         |
| owner_id   | int64           | false  | false    | false    | false   | false         | false     | json:"owner_id,omitempty"   |          0 |         |
| name       | string          | false  | false    | false    | false   | false         | false     | json:"name,omitempty"       |          0 |         |
| created_at | time.Time       | false  | false    | false    | true    | false         | true      | json:"created_at,omitempty" |          0 |         |
| updated_at | time.Time       | false  | false    | false    | true    | false         | false     | json:"updated_at,omitempty" |          0 |         |
| type       | enum.TenantType | false  | false    | false    | true    | false         | true      | json:"type,omitempty"       |          0 |         |


|  Edge   |  Type  | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| members | Member | false   |         | O2M      | false  | true     |         |
| groups  | Group  | false   |         | O2M      | false  | true     |         |
| invites | Invite | false   |         | O2M      | false  | true     |         |

