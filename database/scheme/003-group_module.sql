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
    group_id INT  NOT NULL REFERENCES group_with_policy (id),
    address  TEXT NOT NULL,
    weight   INT  NOT NULL,
    metadata TEXT NULL,
    PRIMARY KEY (group_id, address)
);

CREATE INDEX group_member_weight_index ON group_member (group_id) WHERE weight > 0;
CREATE INDEX group_member_group_id_index ON group_member (group_id);

    
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
    id               INT                         NOT NULL PRIMARY KEY,
    group_id         INT                         NOT NULL REFERENCES group_with_policy (id),
    metadata         TEXT                        NULL,
    proposer         TEXT                        NOT NULL,
    status           PROPOSAL_STATUS             NOT NULL,
    executor_result  PROPOSAL_EXECUTOR_RESULT    NOT NULL,
    executor         TEXT                        NULL,
    execution_time   TIMESTAMP WITHOUT TIME ZONE NULL,
    execution_log    TEXT                        NULL,
    messages         JSONB                       NOT NULL,
    height           BIGINT                      NOT NULL REFERENCES block (height),
    submit_time      TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    transaction_hash TEXT                        NULL REFERENCES transaction (hash)
);

CREATE INDEX group_proposal_status_index ON group_proposal (status);
CREATE INDEX group_proposal_group_id_index ON group_proposal (group_id);
CREATE INDEX group_proposal_submit_time_index ON group_proposal (submit_time);

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
    PRIMARY KEY (proposal_id, voter)
);

CREATE INDEX group_proposal_vote_proposal_id_index ON group_proposal_vote (proposal_id);
