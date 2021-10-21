package accesscontrol

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/stretchr/testify/assert"
)

var testUser = &models.SignedInUser{
	UserId:  2,
	OrgId:   3,
	OrgName: "TestOrg",
	OrgRole: models.ROLE_VIEWER,
	Login:   "testUser",
	Name:    "Test User",
	Email:   "testuser@example.org",
}

func TestResolveKeywordedScope(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.SignedInUser
		permission Permission
		want       *Permission
		wantErr    bool
	}{
		{
			name:       "no scope",
			user:       testUser,
			permission: Permission{Action: "users:read"},
			want:       &Permission{Action: "users:read"},
			wantErr:    false,
		},
		{
			name:       "user if resolution",
			user:       testUser,
			permission: Permission{Action: "users:read", Scope: "users:self"},
			want:       &Permission{Action: "users:read", Scope: "users:id:2"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewScopeResolver()
			resolved, err := resolver.ResolveKeyword(tt.user, tt.permission)
			if tt.wantErr {
				assert.Error(t, err, "expected an error during the resolution of the scope")
				return
			}
			assert.NoError(t, err)
			assert.EqualValues(t, tt.want, resolved, "permission did not match expected resolution")
		})
	}
}

func TestScopeResolver_ResolveAttribute(t *testing.T) {
	testdscmd := &models.AddDataSourceCommand{
		Uid:    "testUID",
		OrgId:  1,
		Name:   "testds",
		Url:    "http://localhost:5432",
		Type:   "postgresql",
		Access: "Proxy",
	}

	tests := []struct {
		name      string
		user      *models.SignedInUser
		initDB    func(t *testing.T, db *sqlstore.SQLStore)
		evaluator Evaluator
		want      Evaluator
		wantErr   bool
	}{
		{
			name:      "nil evaluator",
			user:      nil,
			evaluator: nil,
			want:      nil,
			wantErr:   false,
		},
		{
			name:      "no resolution evaluator",
			user:      nil,
			evaluator: EvalPermission("datasources:read"),
			want:      EvalPermission("datasources:read"),
			wantErr:   false,
		},
		{
			name: "datasource name resolution evaluator",
			user: &models.SignedInUser{OrgId: 1},
			initDB: func(t *testing.T, db *sqlstore.SQLStore) {
				err := db.AddDataSource(context.Background(), testdscmd)
				assert.NoError(t, err)
			},
			evaluator: EvalPermission("datasources:read", Scope("datasources", "name", "testds")),
			want:      EvalPermission("datasources:read", Scope("datasources", "id", "1")),
			wantErr:   false,
		},
		{
			name: "datasource name resolution evaluator",
			user: &models.SignedInUser{OrgId: 1},
			initDB: func(t *testing.T, db *sqlstore.SQLStore) {
				err := db.AddDataSource(context.Background(), testdscmd)
				assert.NoError(t, err)
			},
			evaluator: EvalAll(
				EvalPermission("datasources:read", Scope("datasources", "name", "testds")),
				EvalAny(
					EvalPermission("datasources:read", Scope("datasources", "name", "testds")),
					EvalPermission("datasources:read", Scope("datasources", "name", "testds")),
				),
			),
			want: EvalAll(
				EvalPermission("datasources:read", Scope("datasources", "id", "1")),
				EvalAny(
					EvalPermission("datasources:read", Scope("datasources", "id", "1")),
					EvalPermission("datasources:read", Scope("datasources", "id", "1")),
				),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		db := sqlstore.InitTestDB(t)
		resolver := NewScopeResolver()
		resolver.AddAttributeResolver("datasources:name:", NewResolveDatasourceNameFunc(db))

		if tt.initDB != nil {
			tt.initDB(t, db)
		}

		resolvedEvaluator, err := resolver.ResolveAttribute(context.TODO(), tt.user, tt.evaluator)
		if tt.wantErr {
			assert.Error(t, err, "expected an error during the resolution of the scope")
			return
		}
		assert.NoError(t, err)
		assert.EqualValues(t, tt.want, resolvedEvaluator, "permission did not match expected resolution")
	}
}
