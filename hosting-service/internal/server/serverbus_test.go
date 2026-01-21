package server_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"hosting-kit/page"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"

	"github.com/google/uuid"
)

type mockNotifier struct {
	ServerUpdatedFunc func(ctx context.Context, s server.Server)
}

func (m *mockNotifier) ServerUpdated(ctx context.Context, s server.Server) {
	if m.ServerUpdatedFunc != nil {
		m.ServerUpdatedFunc(ctx, s)
	}
}

type mockResourcesManager struct {
	ConsumeFunc func(ctx context.Context, r server.Resources) (uuid.UUID, error)
	ReturnFunc  func(ctx context.Context, r server.Resources, poolID uuid.UUID) error
}

func (m *mockResourcesManager) Consume(ctx context.Context, r server.Resources) (uuid.UUID, error) {
	if m.ConsumeFunc != nil {
		return m.ConsumeFunc(ctx, r)
	}
	return uuid.New(), nil
}

func (m *mockResourcesManager) Return(ctx context.Context, r server.Resources, poolID uuid.UUID) error {
	if m.ReturnFunc != nil {
		return m.ReturnFunc(ctx, r, poolID)
	}
	return nil
}

type mockPlanFinder struct {
	FindByIDFunc func(ctx context.Context, ID uuid.UUID) (plan.Plan, error)
}

func (m *mockPlanFinder) FindByID(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, ID)
	}
	return plan.Plan{}, nil
}

type mockStorer struct {
	FindByIDFunc func(ctx context.Context, ID uuid.UUID) (server.Server, error)
	CreateFunc   func(ctx context.Context, s server.Server) error
	UpdateFunc   func(ctx context.Context, s server.Server) error
	DeleteFunc   func(ctx context.Context, ID uuid.UUID) error
	FindAllFunc  func(ctx context.Context, pg page.Page, userID uuid.UUID) ([]server.Server, int, error)
}

func (m *mockStorer) FindByID(ctx context.Context, ID uuid.UUID) (server.Server, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, ID)
	}
	return server.Server{}, nil
}
func (m *mockStorer) Create(ctx context.Context, s server.Server) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, s)
	}
	return nil
}
func (m *mockStorer) Update(ctx context.Context, s server.Server) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, s)
	}
	return nil
}
func (m *mockStorer) Delete(ctx context.Context, ID uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, ID)
	}
	return nil
}

func (m *mockStorer) FindAll(ctx context.Context, pg page.Page, userID uuid.UUID) ([]server.Server, int, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, pg, userID)
	}
	return nil, 0, nil
}

type mockProvisioner struct {
	RequestIPFunc func(ctx context.Context, s server.Server) error
}

func (m *mockProvisioner) RequestIP(ctx context.Context, s server.Server) error {
	if m.RequestIPFunc != nil {
		return m.RequestIPFunc(ctx, s)
	}
	return nil
}

func Test_Create(t *testing.T) {
	ctx := context.Background()
	planID := uuid.New()
	userID := uuid.New()
	errBoom := errors.New("boom")

	type testCase struct {
		name       string
		serverName string
		planID     uuid.UUID

		pf   func() *mockPlanFinder
		st   func() *mockStorer
		prov func() *mockProvisioner
		rm   func() *mockResourcesManager

		wantErr error
	}

	table := []testCase{
		{
			name:       "success",
			serverName: "Web01",
			planID:     planID,
			pf: func() *mockPlanFinder {
				return &mockPlanFinder{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
						return plan.Plan{ID: ID, Name: "Basic"}, nil
					},
				}
			},
			st: func() *mockStorer {
				return &mockStorer{
					CreateFunc: func(ctx context.Context, s server.Server) error {
						if s.Status != server.StatusPending {
							return fmt.Errorf("expected status PENDING, got %s", s.Status)
						}
						if s.OwnerID != userID {
							return fmt.Errorf("expected OwnerID %s, got %s", userID, s.OwnerID)
						}
						return nil
					},
				}
			},
			prov:    func() *mockProvisioner { return &mockProvisioner{} },
			rm:      func() *mockResourcesManager { return &mockResourcesManager{} },
			wantErr: nil,
		},
		{
			name:       "fail_plan_not_found",
			serverName: "Web02",
			planID:     planID,
			pf: func() *mockPlanFinder {
				return &mockPlanFinder{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
						return plan.Plan{}, errBoom
					},
				}
			},
			st:      func() *mockStorer { return &mockStorer{} },
			prov:    func() *mockProvisioner { return &mockProvisioner{} },
			rm:      func() *mockResourcesManager { return &mockResourcesManager{} },
			wantErr: errBoom,
		},
		{
			name:       "fail_empty_name",
			serverName: "",
			planID:     planID,
			pf: func() *mockPlanFinder {
				return &mockPlanFinder{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
						return plan.Plan{ID: ID}, nil
					},
				}
			},
			st:      func() *mockStorer { return &mockStorer{} },
			prov:    func() *mockProvisioner { return &mockProvisioner{} },
			rm:      func() *mockResourcesManager { return &mockResourcesManager{} },
			wantErr: server.ErrValidation,
		},
		{
			name:       "fail_provisioner",
			serverName: "Web03",
			planID:     planID,
			pf: func() *mockPlanFinder {
				return &mockPlanFinder{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
						return plan.Plan{ID: ID}, nil
					},
				}
			},
			st: func() *mockStorer { return &mockStorer{} },
			prov: func() *mockProvisioner {
				return &mockProvisioner{
					RequestIPFunc: func(ctx context.Context, s server.Server) error {
						return errBoom
					},
				}
			},
			rm:      func() *mockResourcesManager { return &mockResourcesManager{} },
			wantErr: errBoom,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			bus := server.NewBusiness(tt.st(), tt.pf(), tt.prov(), tt.rm(), &mockNotifier{})

			got, err := bus.Create(ctx, tt.serverName, tt.planID, userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID == uuid.Nil {
				t.Error("expected valid ID, got Nil")
			}
		})
	}
}

func Test_Start(t *testing.T) {
	ctx := context.Background()
	srvID := uuid.New()
	userID := uuid.New()

	type testCase struct {
		name       string
		initStatus server.ServerStatus
		st         func() *mockStorer
		wantErr    error
	}

	table := []testCase{
		{
			name:       "success_from_stopped",
			initStatus: server.StatusStopped,
			st: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{ID: ID, Status: server.StatusStopped, OwnerID: userID}, nil
					},
					UpdateFunc: func(ctx context.Context, s server.Server) error {
						if s.Status != server.StatusRunning {
							return fmt.Errorf("expected RUNNING, got %s", s.Status)
						}
						return nil
					},
				}
			},
			wantErr: nil,
		},
		{
			name:       "fail_wrong_status_pending",
			initStatus: server.StatusPending,
			st: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{ID: ID, Status: server.StatusPending, OwnerID: userID}, nil
					},
				}
			},
			wantErr: server.ErrValidation,
		},
		{
			name:       "fail_not_found",
			initStatus: "",
			st: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{}, server.ErrServerNotFound
					},
				}
			},
			wantErr: server.ErrServerNotFound,
		},
		{
			name:       "fail_access_denied",
			initStatus: server.StatusStopped,
			st: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{ID: ID, Status: server.StatusStopped, OwnerID: uuid.New()}, nil
					},
				}
			},
			wantErr: server.ErrAccessDenied,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			bus := server.NewBusiness(tt.st(), nil, nil, nil, &mockNotifier{})

			_, err := bus.Start(ctx, srvID, userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func Test_SetIPAddress(t *testing.T) {
	ctx := context.Background()
	srvID := uuid.New()
	validIP := "10.0.0.1"

	type testCase struct {
		name        string
		ip          string
		setupStorer func() *mockStorer
		wantErr     error
	}

	table := []testCase{
		{
			name: "success",
			ip:   validIP,
			setupStorer: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{ID: ID, Status: server.StatusPending}, nil
					},
					UpdateFunc: func(ctx context.Context, s server.Server) error {
						if s.Status != server.StatusStopped {
							return fmt.Errorf("expected STOPPED, got %s", s.Status)
						}
						if s.IPv4Address == nil || *s.IPv4Address != validIP {
							return errors.New("ip address mismatch")
						}
						return nil
					},
				}
			},
			wantErr: nil,
		},
		{
			name: "fail_invalid_ip",
			ip:   "not-an-ip",
			setupStorer: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{ID: ID, Status: server.StatusPending}, nil
					},
				}
			},
			wantErr: server.ErrValidation,
		},
		{
			name: "fail_wrong_status_running",
			ip:   validIP,
			setupStorer: func() *mockStorer {
				return &mockStorer{
					FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
						return server.Server{ID: ID, Status: server.StatusRunning}, nil
					},
				}
			},
			wantErr: server.ErrValidation,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			bus := server.NewBusiness(tt.setupStorer(), nil, nil, nil, &mockNotifier{})
			err := bus.SetIPAddress(ctx, srvID, tt.ip)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	ctx := context.Background()
	srvID := uuid.New()
	planID := uuid.New()
	userID := uuid.New()

	type testCase struct {
		name    string
		status  server.ServerStatus
		wantErr error
	}

	table := []testCase{
		{
			name:    "success_stopped",
			status:  server.StatusStopped,
			wantErr: nil,
		},
		{
			name:    "success_running",
			status:  server.StatusRunning,
			wantErr: nil,
		},
		{
			name:    "fail_pending",
			status:  server.StatusPending,
			wantErr: server.ErrValidation,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			st := &mockStorer{
				FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (server.Server, error) {
					return server.Server{ID: ID, Status: tt.status, PlanID: planID, OwnerID: userID}, nil
				},
				DeleteFunc: func(ctx context.Context, ID uuid.UUID) error { return nil },
			}

			pf := &mockPlanFinder{
				FindByIDFunc: func(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
					return plan.Plan{ID: ID}, nil
				},
			}

			rm := &mockResourcesManager{
				ReturnFunc: func(ctx context.Context, r server.Resources, poolID uuid.UUID) error {
					return nil
				},
			}

			bus := server.NewBusiness(st, pf, nil, rm, &mockNotifier{})

			_, err := bus.Delete(ctx, srvID, userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
