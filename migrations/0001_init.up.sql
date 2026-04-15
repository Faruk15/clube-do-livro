-- Membros do clube. Não há auto-cadastro: o admin cria.
CREATE TABLE members (
    id            UUID PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    is_admin      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sessões persistidas (cookie httpOnly carrega o token).
CREATE TABLE sessions (
    id         UUID PRIMARY KEY,
    member_id  UUID        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sessions_member ON sessions(member_id);

-- Livros sugeridos / lidos pelo clube.
-- status: sugerido | em_votacao | lendo_agora | lido
CREATE TABLE books (
    id            UUID PRIMARY KEY,
    title         TEXT        NOT NULL,
    author        TEXT        NOT NULL DEFAULT '',
    cover_url     TEXT        NOT NULL DEFAULT '',
    synopsis      TEXT        NOT NULL DEFAULT '',
    publisher     TEXT        NOT NULL DEFAULT '',
    year          INTEGER     NOT NULL DEFAULT 0,
    pages         INTEGER     NOT NULL DEFAULT 0,
    status        TEXT        NOT NULL DEFAULT 'sugerido',
    suggested_by  UUID            REFERENCES members(id) ON DELETE SET NULL,
    finished_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_books_status ON books(status);

-- Tags livres aplicadas a livros (relação N:N simplificada).
CREATE TABLE book_tags (
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    tag     TEXT NOT NULL,
    PRIMARY KEY (book_id, tag)
);

-- Rodadas de votação. status: aberta | encerrada
CREATE TABLE vote_rounds (
    id              UUID PRIMARY KEY,
    status          TEXT        NOT NULL DEFAULT 'aberta',
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ,
    winner_book_id  UUID            REFERENCES books(id) ON DELETE SET NULL
);

-- Livros candidatos numa rodada.
CREATE TABLE vote_round_books (
    round_id UUID NOT NULL REFERENCES vote_rounds(id) ON DELETE CASCADE,
    book_id  UUID NOT NULL REFERENCES books(id)       ON DELETE CASCADE,
    PRIMARY KEY (round_id, book_id)
);

-- Voto único por membro por rodada.
CREATE TABLE votes (
    id         UUID PRIMARY KEY,
    round_id   UUID        NOT NULL REFERENCES vote_rounds(id) ON DELETE CASCADE,
    member_id  UUID        NOT NULL REFERENCES members(id)     ON DELETE CASCADE,
    book_id    UUID        NOT NULL REFERENCES books(id)       ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (round_id, member_id)
);
CREATE INDEX idx_votes_round_book ON votes(round_id, book_id);

-- Avaliações: todos os campos de nota são opcionais (NULLáveis).
CREATE TABLE reviews (
    id               UUID PRIMARY KEY,
    book_id          UUID        NOT NULL REFERENCES books(id)   ON DELETE CASCADE,
    member_id        UUID        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    nota_geral       SMALLINT,
    nota_escrita     SMALLINT,
    nota_enredo      SMALLINT,
    nota_expectativa SMALLINT,
    review_text      TEXT        NOT NULL DEFAULT '',
    has_spoiler      BOOLEAN     NOT NULL DEFAULT FALSE,
    citacao          TEXT        NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (book_id, member_id)
);
CREATE INDEX idx_reviews_book ON reviews(book_id);

-- Encontros do clube.
CREATE TABLE meetings (
    id          UUID PRIMARY KEY,
    title       TEXT        NOT NULL,
    datetime    TIMESTAMPTZ NOT NULL,
    location    TEXT        NOT NULL DEFAULT '',
    book_id     UUID            REFERENCES books(id) ON DELETE SET NULL,
    reminded_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_meetings_datetime ON meetings(datetime);

-- Confirmação de presença. status: confirmado | nao_vou | talvez
CREATE TABLE attendances (
    id         UUID PRIMARY KEY,
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    member_id  UUID NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
    status     TEXT NOT NULL DEFAULT 'confirmado',
    UNIQUE (meeting_id, member_id)
);

-- Pauta colaborativa por encontro.
CREATE TABLE agenda_items (
    id         UUID PRIMARY KEY,
    meeting_id UUID        NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    member_id  UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_agenda_meeting ON agenda_items(meeting_id);
