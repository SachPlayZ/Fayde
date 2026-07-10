package boards_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/boards"
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

// makeFriends creates a mutual friendship between userA and userB.
func makeFriends(t *testing.T, userA, userBEmail string) {
	t.Helper()
	ctx := context.Background()
	friendsSvc := friends.NewService(friends.NewRepository(testPool))
	f, err := friendsSvc.SendRequest(ctx, userA, userBEmail)
	require.NoError(t, err)
	// Find userB ID.
	repo := auth.NewRepository(testPool)
	userB, err := repo.GetUserByEmail(ctx, userBEmail)
	require.NoError(t, err)
	require.NoError(t, friendsSvc.AcceptRequest(ctx, userB.ID, f.ID))
}

func newService() *boards.Service {
	return boards.NewService(boards.NewRepository(testPool))
}

// TestCreateBoard verifies a board is created and the owner is a member.
func TestCreateBoard(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "board_create_owner@example.com")

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "My Board", Description: "Desc"})
	require.NoError(t, err)
	assert.Equal(t, "My Board", b.Name)
	assert.Equal(t, ownerID, b.OwnerID)

	// Owner should be able to get the board detail.
	detail, err := svc.GetBoard(ctx, ownerID, b.ID)
	require.NoError(t, err)
	require.Len(t, detail.Members, 1)
	assert.Equal(t, "owner", detail.Members[0].Role)
}

// TestAddBoardTask verifies a member can add a task.
func TestAddBoardTask(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "task_owner@example.com")
	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Task Board"})
	require.NoError(t, err)

	task, err := svc.AddTask(ctx, ownerID, b.ID, boards.AddTaskInput{Title: "Go to the Gym"})
	require.NoError(t, err)
	assert.Equal(t, "Go to the Gym", task.Title)
	assert.Equal(t, b.ID, task.BoardID)
}

// TestCompleteUncompleteIdempotency verifies completing twice doesn't error or double-insert.
func TestCompleteUncompleteIdempotency(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "idem_owner@example.com")
	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Idem Board"})
	require.NoError(t, err)

	task, err := svc.AddTask(ctx, ownerID, b.ID, boards.AddTaskInput{Title: "Meditate"})
	require.NoError(t, err)

	// Complete once.
	c1, err := svc.Complete(ctx, ownerID, b.ID, task.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, c1.ID)

	// Complete again (same day) — should not error, ON CONFLICT DO UPDATE.
	c2, err := svc.Complete(ctx, ownerID, b.ID, task.ID)
	require.NoError(t, err)
	assert.Equal(t, c1.CompletionDate, c2.CompletionDate)

	// Uncomplete.
	require.NoError(t, svc.Uncomplete(ctx, ownerID, b.ID, task.ID))

	// Uncomplete again — should not error.
	require.NoError(t, svc.Uncomplete(ctx, ownerID, b.ID, task.ID))

	// Verify completion is gone.
	detail, err := svc.GetBoard(ctx, ownerID, b.ID)
	require.NoError(t, err)
	assert.Empty(t, detail.Completions)
}

// TestJoinViaShareToken verifies the full join flow.
func TestJoinViaShareToken(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "join_owner@example.com")
	joinerID := createTestUser(t, "join_joiner@example.com")

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Join Board"})
	require.NoError(t, err)

	// Create invite token.
	inv, err := svc.CreateShareToken(ctx, ownerID, b.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, inv.Token)

	// Preview (public, no auth).
	preview, err := svc.JoinPreview(ctx, inv.Token)
	require.NoError(t, err)
	assert.Equal(t, "Join Board", preview.BoardName)
	assert.Equal(t, 1, preview.MemberCount)

	// Join via token.
	joined, err := svc.JoinViaToken(ctx, joinerID, inv.Token)
	require.NoError(t, err)
	assert.Equal(t, b.ID, joined.ID)

	// Joiner should now be a member.
	detail, err := svc.GetBoard(ctx, joinerID, b.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Members, 2)
}

// TestJoinAlreadyMember verifies joining when already a member doesn't error.
func TestJoinAlreadyMember(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "join_idem_owner@example.com")
	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Already Member Board"})
	require.NoError(t, err)

	inv, err := svc.CreateShareToken(ctx, ownerID, b.ID)
	require.NoError(t, err)

	// Owner joins again via token — should not error.
	_, err = svc.JoinViaToken(ctx, ownerID, inv.Token)
	require.NoError(t, err)

	// Still only one member.
	detail, err := svc.GetBoard(ctx, ownerID, b.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Members, 1)
}

// TestNonMemberCannotSeeBoard verifies a non-member gets ErrNotMember on GetBoard.
func TestNonMemberCannotSeeBoard(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "nonmem_owner@example.com")
	outsiderID := createTestUser(t, "nonmem_outsider@example.com")

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Private Board"})
	require.NoError(t, err)

	_, err = svc.GetBoard(ctx, outsiderID, b.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, boards.ErrNotMember))
}

// TestNonMemberCannotComplete verifies a non-member cannot complete a board task.
func TestNonMemberCannotComplete(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "nmcomp_owner@example.com")
	outsiderID := createTestUser(t, "nmcomp_outsider@example.com")

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Comp Board"})
	require.NoError(t, err)

	task, err := svc.AddTask(ctx, ownerID, b.ID, boards.AddTaskInput{Title: "Run"})
	require.NoError(t, err)

	_, err = svc.Complete(ctx, outsiderID, b.ID, task.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, boards.ErrNotMember))
}

// TestOwnerOnlyDelete verifies only the owner can delete a board.
func TestOwnerOnlyDelete(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "del_owner@example.com")
	memberEmail := "del_member@example.com"
	createTestUser(t, memberEmail)

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Delete Board"})
	require.NoError(t, err)

	// Make the other user a friend first, then invite.
	makeFriends(t, ownerID, memberEmail)
	repo := auth.NewRepository(testPool)
	member, err := repo.GetUserByEmail(ctx, memberEmail)
	require.NoError(t, err)

	require.NoError(t, svc.InviteFriend(ctx, ownerID, b.ID, member.ID))

	// Member tries to delete — forbidden.
	err = svc.DeleteBoard(ctx, member.ID, b.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, boards.ErrForbidden))

	// Owner deletes — succeeds.
	require.NoError(t, svc.DeleteBoard(ctx, ownerID, b.ID))
}

// TestInviteFriendNotFriends verifies non-friends cannot be directly invited.
func TestInviteFriendNotFriends(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "invfr_owner@example.com")
	strangerID := createTestUser(t, "invfr_stranger@example.com")

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Invite Board"})
	require.NoError(t, err)

	err = svc.InviteFriend(ctx, ownerID, b.ID, strangerID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, boards.ErrNotFriends))
}

// TestRevokeShareToken verifies the share token can be revoked and join fails afterwards.
func TestRevokeShareToken(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "revoke_owner@example.com")
	joinerID := createTestUser(t, "revoke_joiner@example.com")

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "Revoke Board"})
	require.NoError(t, err)

	inv, err := svc.CreateShareToken(ctx, ownerID, b.ID)
	require.NoError(t, err)

	// Revoke.
	require.NoError(t, svc.RevokeShareToken(ctx, ownerID, b.ID))

	// Join attempt should fail.
	_, err = svc.JoinViaToken(ctx, joinerID, inv.Token)
	require.Error(t, err)
	assert.True(t, errors.Is(err, boards.ErrInvalidToken))
}

// TestDeleteTaskOwnerOrCreatorOnly verifies only the owner or creator can delete a task.
func TestDeleteTaskOwnerOrCreatorOnly(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	ownerID := createTestUser(t, "deltask_owner@example.com")
	memberEmail := "deltask_member@example.com"
	createTestUser(t, memberEmail)

	b, err := svc.CreateBoard(ctx, ownerID, boards.CreateBoardInput{Name: "DelTask Board"})
	require.NoError(t, err)

	makeFriends(t, ownerID, memberEmail)
	authRepo := auth.NewRepository(testPool)
	member, err := authRepo.GetUserByEmail(ctx, memberEmail)
	require.NoError(t, err)
	require.NoError(t, svc.InviteFriend(ctx, ownerID, b.ID, member.ID))

	// Task created by owner.
	task, err := svc.AddTask(ctx, ownerID, b.ID, boards.AddTaskInput{Title: "Owner Task"})
	require.NoError(t, err)

	// Member tries to delete — forbidden.
	err = svc.DeleteTask(ctx, member.ID, b.ID, task.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, boards.ErrForbidden))

	// Owner deletes — succeeds.
	require.NoError(t, svc.DeleteTask(ctx, ownerID, b.ID, task.ID))
}
