-- Source: github.com/marcus/td v0.37.0 internal/db/schema.go BaseSchema (MIT).

-- Issues table
CREATE TABLE IF NOT EXISTS issues (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    type TEXT NOT NULL DEFAULT 'task',
    priority TEXT NOT NULL DEFAULT 'P2',
    points INTEGER DEFAULT 0,
    labels TEXT DEFAULT '',
    parent_id TEXT DEFAULT '',
    acceptance TEXT DEFAULT '',
    implementer_session TEXT DEFAULT '',
    reviewer_session TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    deleted_at DATETIME,
    minor INTEGER DEFAULT 0,
    created_branch TEXT DEFAULT '',
    FOREIGN KEY (parent_id) REFERENCES issues(id)
);

-- Logs table
CREATE TABLE IF NOT EXISTS logs (
    id TEXT PRIMARY KEY,
    issue_id TEXT DEFAULT '',
    session_id TEXT NOT NULL,
    work_session_id TEXT DEFAULT '',
    message TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'progress',
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Handoffs table
CREATE TABLE IF NOT EXISTS handoffs (
    id TEXT PRIMARY KEY,
    issue_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    done TEXT DEFAULT '[]',
    remaining TEXT DEFAULT '[]',
    decisions TEXT DEFAULT '[]',
    uncertain TEXT DEFAULT '[]',
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (issue_id) REFERENCES issues(id)
);

-- Git snapshots table
CREATE TABLE IF NOT EXISTS git_snapshots (
    id TEXT PRIMARY KEY,
    issue_id TEXT NOT NULL,
    event TEXT NOT NULL,
    commit_sha TEXT NOT NULL,
    branch TEXT NOT NULL,
    dirty_files INTEGER DEFAULT 0,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (issue_id) REFERENCES issues(id)
);

-- Issue files table
CREATE TABLE IF NOT EXISTS issue_files (
    id TEXT PRIMARY KEY,
    issue_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'implementation',
    linked_sha TEXT DEFAULT '',
    linked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (issue_id) REFERENCES issues(id),
    UNIQUE(issue_id, file_path)
);

-- Issue dependencies table
CREATE TABLE IF NOT EXISTS issue_dependencies (
    id TEXT PRIMARY KEY,
    issue_id TEXT NOT NULL,
    depends_on_id TEXT NOT NULL,
    relation_type TEXT NOT NULL DEFAULT 'depends_on',
    UNIQUE(issue_id, depends_on_id, relation_type),
    FOREIGN KEY (issue_id) REFERENCES issues(id),
    FOREIGN KEY (depends_on_id) REFERENCES issues(id)
);

-- Work sessions table
CREATE TABLE IF NOT EXISTS work_sessions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    session_id TEXT NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    start_sha TEXT DEFAULT '',
    end_sha TEXT DEFAULT ''
);

-- Work session issues junction table
CREATE TABLE IF NOT EXISTS work_session_issues (
    id TEXT PRIMARY KEY,
    work_session_id TEXT NOT NULL,
    issue_id TEXT NOT NULL,
    tagged_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(work_session_id, issue_id),
    FOREIGN KEY (work_session_id) REFERENCES work_sessions(id),
    FOREIGN KEY (issue_id) REFERENCES issues(id)
);

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id TEXT PRIMARY KEY,
    issue_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (issue_id) REFERENCES issues(id)
);

-- Sessions table for tracking session history
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    name TEXT DEFAULT '',
    branch TEXT DEFAULT '',
    agent_type TEXT DEFAULT '',
    agent_pid INTEGER DEFAULT 0,
    context_id TEXT DEFAULT '',
    previous_session_id TEXT DEFAULT '',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    last_activity DATETIME
);

-- Schema info table for version tracking
CREATE TABLE IF NOT EXISTS schema_info (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_priority ON issues(priority);
CREATE INDEX IF NOT EXISTS idx_issues_type ON issues(type);
CREATE INDEX IF NOT EXISTS idx_issues_parent ON issues(parent_id);
CREATE INDEX IF NOT EXISTS idx_issues_deleted ON issues(deleted_at);
CREATE INDEX IF NOT EXISTS idx_logs_issue ON logs(issue_id);
CREATE INDEX IF NOT EXISTS idx_handoffs_issue ON handoffs(issue_id);
CREATE INDEX IF NOT EXISTS idx_git_snapshots_issue ON git_snapshots(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_files_issue ON issue_files(issue_id);
CREATE INDEX IF NOT EXISTS idx_comments_issue ON comments(issue_id);
CREATE INDEX IF NOT EXISTS idx_sessions_branch ON sessions(branch);
CREATE INDEX IF NOT EXISTS idx_sessions_branch_agent ON sessions(branch, agent_type, agent_pid);
