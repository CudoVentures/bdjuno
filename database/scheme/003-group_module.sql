CREATE TABLE group_with_policy
(
    id                   INT    NOT NULL PRIMARY KEY,
    address              TEXT   NOT NULL,
    group_metadata       TEXT   NULL,
    policy_metadata      TEXT   NULL,
    threshold            INT    NOT NULL,
    voting_period        BIGINT NOT NULL,
    min_execution_period BIGINT NOT NULL
);

CREATE TABLE group_member
(
    group_id        INT  NOT NULL REFERENCES group_with_policy (id),
    address         TEXT NOT NULL,
    weight          INT  NOT NULL,
    member_metadata TEXT NULL,
    PRIMARY KEY(group_id, address)
);

CREATE TYPE PROPOSAL_STATUS AS ENUM
(
    'PROPOSAL_STATUS_SUBMITTED',
    'PROPOSAL_STATUS_ACCEPTED',
    'PROPOSAL_STATUS_REJECTED',
    'PROPOSAL_STATUS_ABORTED',
    'PROPOSAL_STATUS_WITHDRAWN'
);

CREATE TYPE PROPOSAL_EXECUTOR_RESULT AS ENUM 
(
    'PROPOSAL_EXECUTOR_RESULT_NOT_RUN',
    'PROPOSAL_EXECUTOR_RESULT_SUCCESS',
    'PROPOSAL_EXECUTOR_RESULT_FAILURE'
);

CREATE TABLE group_proposal
(
    id                INT                         NOT NULL PRIMARY KEY,
    group_id          INT                         NOT NULL REFERENCES group_with_policy (id),
    proposal_metadata TEXT                        NULL,
    proposer          TEXT                        NOT NULL,
    status            PROPOSAL_STATUS             NOT NULL,
    executor_result   PROPOSAL_EXECUTOR_RESULT    NOT NULL,
    messages          JSONB                       NOT NULL DEFAULT '{}'::JSONB,
    height            BIGINT                      NOT NULL REFERENCES block (height),
    transaction_hash  TEXT                        NULL REFERENCES transaction (hash)
);

CREATE TYPE VOTE_OPTION AS ENUM
(
    'VOTE_OPTION_YES',
    'VOTE_OPTION_ABSTAIN',
    'VOTE_OPTION_NO',
    'VOTE_OPTION_NO_WITH_VETO'
);

CREATE TABLE group_proposal_vote
(
    proposal_id   INT                         NOT NULL REFERENCES group_proposal (id),
    group_id      INT                         NOT NULL,
    voter         TEXT                        NOT NULL,
    vote_option   VOTE_OPTION                 NOT NULL,
    vote_metadata TEXT                        NULL,
    submit_time   TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    FOREIGN KEY (group_id, voter) REFERENCES group_member (group_id, address),
    PRIMARY KEY(proposal_id, voter)
);