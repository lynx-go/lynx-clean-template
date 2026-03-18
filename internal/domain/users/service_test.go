package users

import (
	"context"
	"testing"

	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type fakePasswordHasher struct{}

func (f fakePasswordHasher) Hash(password string) (string, error) {
	return "hashed-" + password, nil
}

func (f fakePasswordHasher) Compare(_, _ string) error {
	return nil
}

type fakeUsersRepo struct {
	usersByUsername map[string]*repo.User
	usersByEmail    map[string]*repo.User
	createdUsers    []*repo.User
	firstUser       bool
}

func (f *fakeUsersRepo) CreateWithFirstUserSuperAdmin(_ context.Context, u *repo.User) (bool, error) {
	if f.firstUser {
		u.IsSuperAdmin = true
		u.Role = "admin"
	}
	f.createdUsers = append(f.createdUsers, u)
	return f.firstUser, nil
}

func (f *fakeUsersRepo) GetByUsername(_ context.Context, username string) (*repo.User, error) {
	if f.usersByUsername == nil {
		return nil, nil
	}
	return f.usersByUsername[username], nil
}

func (f *fakeUsersRepo) GetByEmail(_ context.Context, email string) (*repo.User, error) {
	if f.usersByEmail == nil {
		return nil, nil
	}
	return f.usersByEmail[email], nil
}

func (f *fakeUsersRepo) IsSuperAdmin(_ context.Context, _ idgen.ID) (bool, error) {
	return false, nil
}

func (f *fakeUsersRepo) SetSuperAdmin(_ context.Context, _ idgen.ID, _ bool) error {
	return nil
}

func (f *fakeUsersRepo) Create(_ context.Context, _ *repo.User) error { return nil }
func (f *fakeUsersRepo) Update(_ context.Context, _ *repo.User, _ ...string) error { return nil }
func (f *fakeUsersRepo) Delete(_ context.Context, _ *repo.User) error { return nil }
func (f *fakeUsersRepo) List(_ context.Context, _ crud.ListParams) ([]*repo.User, int, string, error) {
	return nil, 0, "", nil
}
func (f *fakeUsersRepo) Get(_ context.Context, _ idgen.ID) (*repo.User, error) { return nil, nil }
func (f *fakeUsersRepo) BatchGet(_ context.Context, _ []idgen.ID) ([]*repo.User, error) {
	return nil, nil
}
func (f *fakeUsersRepo) BatchDelete(_ context.Context, _ []idgen.ID) error { return nil }

func newTestUserService(repo *fakeUsersRepo) *UserService {
	return NewUserService(repo, fakePasswordHasher{}, &config.AppConfig{
		App: &config.App{
			User: &config.User{DefaultAvatarUrls: []string{"https://example.com/avatar.png"}},
		},
	})
}

func TestCreate_FirstUserBecomesSuperAdmin(t *testing.T) {
	userRepo := &fakeUsersRepo{firstUser: true}
	svc := newTestUserService(userRepo)

	id, err := svc.Create(context.Background(), &UserCreateRequest{
		Username:    "first",
		DisplayName: "first",
		Email:       "first@example.com",
		Password:    "1234567",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if id == "" {
		t.Fatal("Create() returned empty id")
	}
	if len(userRepo.createdUsers) != 1 {
		t.Fatalf("created user count = %d, want 1", len(userRepo.createdUsers))
	}
	created := userRepo.createdUsers[0]
	if !created.IsSuperAdmin {
		t.Fatal("first user should be super admin")
	}
	if created.Role != "admin" {
		t.Fatalf("first user role = %q, want admin", created.Role)
	}
}

func TestCreate_NonFirstUserStaysNormalUser(t *testing.T) {
	userRepo := &fakeUsersRepo{firstUser: false}
	svc := newTestUserService(userRepo)

	_, err := svc.Create(context.Background(), &UserCreateRequest{
		Username:    "second",
		DisplayName: "second",
		Email:       "second@example.com",
		Password:    "1234567",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(userRepo.createdUsers) != 1 {
		t.Fatalf("created user count = %d, want 1", len(userRepo.createdUsers))
	}
	created := userRepo.createdUsers[0]
	if created.IsSuperAdmin {
		t.Fatal("non-first user should not be super admin")
	}
	if created.Role != "user" {
		t.Fatalf("non-first user role = %q, want user", created.Role)
	}
}

