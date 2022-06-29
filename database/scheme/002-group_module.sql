-- Active: 1656007054073@@127.0.0.1@5432@bdjuno1@public
CREATE TABLE group_with_policy
(
    id                    INTEGER  NOT NULL PRIMARY KEY,
    address               TEXT     NOT NULL,
    group_metadata        TEXT     NULL,
    policy_metadata       TEXT     NULL,
    threshold             INT      NOT NULL CHECK (threshold > 0),
    voting_period         BIGINT   NOT NULL CHECK (voting_period > 0),
    min_execution_period  BIGINT   NOT NULL CHECK (min_execution_period >= 0)
);

CREATE TABLE group_member
(
    group_id         INTEGER  NOT NULL REFERENCES group_with_policy (id),
    address          TEXT     NOT NULL,
    weight           INT      NOT NULL CHECK (weight > 0),
    member_metadata  TEXT     NULL,
    PRIMARY KEY(group_id, address)
);

CREATE TYPE PROPOSAL_STATUS AS ENUM
(
    'PROPOSAL_STATUS_UNSPECIFIED',
    'PROPOSAL_STATUS_SUBMITTED',
    'PROPOSAL_STATUS_ACCEPTED',
    'PROPOSAL_STATUS_REJECTED',
    'PROPOSAL_STATUS_ABORTED'
);

CREATE TYPE PROPOSAL_EXEUTOR_RESULT AS ENUM 
(
    'PROPOSAL_EXECUTOR_RESULT_UNSPECIFIED',
    'PROPOSAL_EXECUTOR_RESULT_NOT_RUN',
    'PROPOSAL_EXECUTOR_RESULT_SUCCESS',
    'PROPOSAL_EXECUTOR_RESULT_FAILURE'
);

CREATE TABLE group_proposal
(
    id                 INTEGER                  NOT NULL PRIMARY KEY,
    group_id           INTEGER                  NOT NULL REFERENCES group_with_policy (id),
    proposal_metadata  TEXT                     NULL,
    proposers          TEXT[]                   NOT NULL,
    submit_time        TIMESTAMP                WITHOUT TIME ZONE NOT NULL,
    status             PROPOSAL_STATUS          NOT NULL,
    executor_result    PROPOSAL_EXEUTOR_RESULT  NOT NULL,
    messages           JSONB                    NOT NULL DEFAULT '{}'::JSONB
);

CREATE TYPE VOTE_OPTION AS ENUM
(
    'VOTE_OPTION_UNSPECIFIED',
    'VOTE_OPTION_YES',
    'VOTE_OPTION_ABSTAIN',
    'VOTE_OPTION_NO',
    'VOTE_OPTION_NO_WITH_VETO'
);

CREATE TABLE group_proposal_vote
(
    id             INTEGER      NOT NULL PRIMARY KEY,
    proposal_id    INTEGER      NOT NULL REFERENCES group_proposal (id),
    voter          TEXT         NOT NULL REFERENCES account (address),
    vote_option    VOTE_OPTION  NOT NULL,
    vote_metadata  TEXT         NULL,
    submit_time    TIMESTAMP    WITHOUT TIME ZONE NOT NULL
);