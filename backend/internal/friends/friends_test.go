package friends_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/db"
	"github.com/SachPlayZ/rivz-asn/backend/internal/friends"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	var connStr string

	if testURL := os.Getenv("TEST_DATABASE_URL"); testURL != "" {
		connStr = testURL
	} else {
		pgContainer, err := postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("testuser"),
			postgres.WithPassword("testpass"),
			postgres.BasicWaitStrategies(),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "start postgres container: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := pgContainer.Terminate(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "terminate container: %v\n", err)
			}
		}()

		connStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			fmt.Fprintf(os.Stderr, "get connection string: %v\n", err)
			os.Exit(1)
		}
	}

	var err error
	testPool, err = db.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect pool: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	migrateURL := toPgx5URL(connStr)
	if err := db.RunMigrations(migrateURL); err != nil && err.Error() != "no change" {
		fmt.Fprintf(os.Stderr, "run migrations: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func toPgx5URL(u string) string {
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if len(u) > len(prefix) && u[:len(prefix)] == prefix {
			return "pgx5://" + u[len(prefix):]
		}
	}
	return u
}

func createTestUser(t *testing.T, email string) string {
	t.Helper()
	repo := auth.NewRepository(testPool)
	svc := auth.NewService(repo, "test-secret", nil, "", nil)
	err := svc.Signup(context.Background(), email, "password123", "Test User")
	require.NoError(t, err)
	user, err := repo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	return user.ID
}

func newService() *friends.Service {
	return friends.NewService(friends.NewRepository(testPool))
}

// TestSendRequest verifies a friend request can be sent.
func TestSendRequest(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "send_a@example.com")
	createTestUser(t, "send_b@example.com")

	f, err := svc.SendRequest(ctx, userA, "send_b@example.com")
	require.NoError(t, err)
	assert.Equal(t, "pending", f.Status)
	assert.Equal(t, userA, f.RequesterID)
}

// TestDuplicateRequestRejection verifies a second request to the same user is rejected.
func TestDuplicateRequestRejection(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "dup_a@example.com")
	createTestUser(t, "dup_b@example.com")

	_, err := svc.SendRequest(ctx, userA, "dup_b@example.com")
	require.NoError(t, err)

	_, err = svc.SendRequest(ctx, userA, "dup_b@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, friends.ErrAlreadyExists))
}

// TestSelfAddRejection verifies cannot add yourself.
func TestSelfAddRejection(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "self_a@example.com")

	_, err := svc.SendRequest(ctx, userA, "self_a@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, friends.ErrCannotSelfAdd))
}

// TestAcceptRequest verifies accepting transitions status to accepted.
func TestAcceptRequest(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "accept_a@example.com")
	userB := createTestUser(t, "accept_b@example.com")

	f, err := svc.SendRequest(ctx, userA, "accept_b@example.com")
	require.NoError(t, err)

	err = svc.AcceptRequest(ctx, userB, f.ID)
	require.NoError(t, err)

	// Friendship should appear in list.
	list, err := svc.ListFriends(ctx, userA)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, userB, list[0].User.ID)
}

// TestDeclineRequest verifies declining transitions to declined and removes from friends list.
func TestDeclineRequest(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "decline_a@example.com")
	userB := createTestUser(t, "decline_b@example.com")

	f, err := svc.SendRequest(ctx, userA, "decline_b@example.com")
	require.NoError(t, err)

	err = svc.DeclineRequest(ctx, userB, f.ID)
	require.NoError(t, err)

	// Should not appear in friends list.
	list, err := svc.ListFriends(ctx, userA)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestDeclineForbiddenForRequester verifies the requester cannot decline their own sent request.
func TestDeclineForbiddenForRequester(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "decfb_a@example.com")
	createTestUser(t, "decfb_b@example.com")

	f, err := svc.SendRequest(ctx, userA, "decfb_b@example.com")
	require.NoError(t, err)

	// userA is the requester, cannot decline (only addressee can).
	err = svc.DeclineRequest(ctx, userA, f.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, friends.ErrForbidden))
}

// TestListRequests verifies pending requests appear in ListRequests with correct direction.
func TestListRequests(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "listreq_a@example.com")
	userB := createTestUser(t, "listreq_b@example.com")

	_, err := svc.SendRequest(ctx, userA, "listreq_b@example.com")
	require.NoError(t, err)

	// From A's perspective: outgoing.
	reqsA, err := svc.ListRequests(ctx, userA)
	require.NoError(t, err)
	require.Len(t, reqsA, 1)
	assert.Equal(t, "outgoing", reqsA[0].Direction)
	assert.Equal(t, userB, reqsA[0].User.ID)

	// From B's perspective: incoming.
	reqsB, err := svc.ListRequests(ctx, userB)
	require.NoError(t, err)
	require.Len(t, reqsB, 1)
	assert.Equal(t, "incoming", reqsB[0].Direction)
	assert.Equal(t, userA, reqsB[0].User.ID)
}

// TestRemoveFriend verifies removing a friend deletes the relationship.
func TestRemoveFriend(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	userA := createTestUser(t, "remfr_a@example.com")
	userB := createTestUser(t, "remfr_b@example.com")

	f, err := svc.SendRequest(ctx, userA, "remfr_b@example.com")
	require.NoError(t, err)
	require.NoError(t, svc.AcceptRequest(ctx, userB, f.ID))

	// Both should see each other as friends.
	list, err := svc.ListFriends(ctx, userA)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Remove.
	require.NoError(t, svc.Remove(ctx, userA, f.ID))

	list, err = svc.ListFriends(ctx, userA)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestSearchUsers verifies email prefix search works.
func TestSearchUsers(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	createTestUser(t, "search_alpha@example.com")
	createTestUser(t, "search_beta@example.com")

	results, err := svc.SearchUsers(ctx, "search_alpha")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "search_alpha@example.com", results[0].Email)

	// Empty query returns empty slice.
	results, err = svc.SearchUsers(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, results)
}
