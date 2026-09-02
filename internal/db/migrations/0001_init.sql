-- Blog posts. body_md is the source of truth; body_html is rendered once at
-- save time so the public read path never parses Markdown per request.
CREATE TABLE posts (
    id           INTEGER  PRIMARY KEY AUTOINCREMENT,
    slug         TEXT     NOT NULL UNIQUE,
    title        TEXT     NOT NULL,
    summary      TEXT     NOT NULL DEFAULT '',
    body_md      TEXT     NOT NULL DEFAULT '',
    body_html    TEXT     NOT NULL DEFAULT '',
    published    INTEGER  NOT NULL DEFAULT 0,
    published_at DATETIME,
    created_at   DATETIME NOT NULL,
    updated_at   DATETIME NOT NULL
);

CREATE INDEX idx_posts_published ON posts (published, published_at DESC);

-- Portfolio projects. Same shape as posts plus presentation metadata.
CREATE TABLE projects (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    slug       TEXT     NOT NULL UNIQUE,
    title      TEXT     NOT NULL,
    summary    TEXT     NOT NULL DEFAULT '',
    body_md    TEXT     NOT NULL DEFAULT '',
    body_html  TEXT     NOT NULL DEFAULT '',
    role       TEXT     NOT NULL DEFAULT '',
    tech       TEXT     NOT NULL DEFAULT '',
    link_url   TEXT     NOT NULL DEFAULT '',
    repo_url   TEXT     NOT NULL DEFAULT '',
    sort_order INTEGER  NOT NULL DEFAULT 0,
    published  INTEGER  NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX idx_projects_published ON projects (published, sort_order, created_at DESC);
