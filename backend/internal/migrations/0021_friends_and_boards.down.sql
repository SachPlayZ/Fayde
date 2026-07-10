DROP INDEX IF EXISTS idx_board_invites_token;
DROP INDEX IF EXISTS idx_board_invites_board;
DROP TABLE IF EXISTS board_invites;

DROP INDEX IF EXISTS idx_board_task_completions_user;
DROP INDEX IF EXISTS idx_board_task_completions_task;
DROP TABLE IF EXISTS board_task_completions;

DROP INDEX IF EXISTS idx_board_tasks_board;
DROP TABLE IF EXISTS board_tasks;

DROP INDEX IF EXISTS idx_board_members_user;
DROP TABLE IF EXISTS board_members;

DROP INDEX IF EXISTS idx_boards_owner;
DROP TABLE IF EXISTS boards;

DROP INDEX IF EXISTS idx_friendships_addressee;
DROP INDEX IF EXISTS idx_friendships_requester;
DROP TABLE IF EXISTS friendships;
