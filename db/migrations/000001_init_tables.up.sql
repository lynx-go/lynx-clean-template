--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users
(
    id                   character varying(64)                  NOT NULL,
    username             character varying(255)                 NOT NULL,
    display_name         character varying(255)                 NOT NULL,
    password_hash        character varying(255),
    avatar_url           character varying(512),
    phone                character varying(32),
    phone_confirmed_at   timestamp with time zone,
    email                character varying(256),
    email_confirmed_at   timestamp with time zone,
    status               integer                  DEFAULT 0     NOT NULL,
    gender               integer                  DEFAULT 0     NOT NULL,
    app_metadata         jsonb,
    user_metadata        jsonb,
    last_sign_in_at      timestamp with time zone DEFAULT now() NOT NULL,
    banned_until         timestamp with time zone,
    confirmed_at         timestamp with time zone,
    confirmation_token   character varying(255),
    confirmation_sent_at timestamp with time zone,
    recovery_token       character varying(255),
    recovery_sent_at     timestamp with time zone,
    role                 character varying(255),
    is_super_admin       boolean                  DEFAULT false,
    created_at           timestamp with time zone DEFAULT now() NOT NULL,
    updated_at           timestamp with time zone DEFAULT now() NOT NULL,
    created_by           character varying(64)                  NOT NULL,
    updated_by           character varying(64)                  NOT NULL
);



--
-- Name: users users_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pk PRIMARY KEY (id);


--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON TABLE public.users IS '用户表';


--
-- Name: COLUMN users.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.id IS '用户主键 (ULID)';


--
-- Name: COLUMN users.username; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.username IS '用户名';


--
-- Name: COLUMN users.display_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.display_name IS '显示用户名';


--
-- Name: COLUMN users.password_hash; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.password_hash IS '密码';


--
-- Name: COLUMN users.avatar_url; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.avatar_url IS '头像';


--
-- Name: COLUMN users.phone_confirmed_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.phone_confirmed_at IS '手机号验证时间戳';


--
-- Name: COLUMN users.email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.email IS '邮箱';


--
-- Name: COLUMN users.email_confirmed_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.email_confirmed_at IS '邮箱验证时间戳';


--
-- Name: COLUMN users.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.status IS '状态，0-初始（未验证），1-正常（已验证），-1-被锁定，-2-已注销';


--
-- Name: COLUMN users.gender; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.gender IS '性别，0-未设置，1-男，2-女';


--
-- Name: COLUMN users.app_metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.app_metadata IS '应用元数据';


--
-- Name: COLUMN users.user_metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.user_metadata IS '用户元数据';


--
-- Name: COLUMN users.last_sign_in_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.last_sign_in_at IS '最近登录时间戳';


--
-- Name: COLUMN users.banned_until; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.banned_until IS '封禁到期时间戳';


--
-- Name: COLUMN users.confirmed_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.confirmed_at IS '验证时间戳';


--
-- Name: COLUMN users.role; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.role IS '角色';


--
-- Name: COLUMN users.created_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.created_at IS '创建时间戳';


--
-- Name: COLUMN users.updated_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.users.updated_at IS '更新时间戳';

--
-- Name: groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.groups
(
    id           character varying(64)                  NOT NULL,
    name         character varying                      NOT NULL,
    display_name character varying                      NOT NULL,
    owner_id     character varying(64)                  NOT NULL,
    icon         character varying,
    description  character varying,
    plan_id      character varying        DEFAULT 0     NOT NULL,
    status       integer                  DEFAULT 1     NOT NULL,
    type         character varying        DEFAULT 'custom'::character varying NOT NULL,
    created_at   timestamp with time zone DEFAULT now() NOT NULL,
    updated_at   timestamp with time zone DEFAULT now() NOT NULL,
    created_by   character varying(64)                  NOT NULL,
    updated_by   character varying(64)                  NOT NULL
);


--
-- Name: TABLE groups; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON TABLE public.groups IS '团队';


--
-- Name: COLUMN groups.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.groups.id IS '主键 (ULID)';


--
-- Name: COLUMN groups.plan_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.groups.plan_id IS '付费计划';


--
-- Name: COLUMN groups.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.groups.type IS '团队类型: system(系统预定义), personal(个人默认), custom(普通团队)';

--
-- Name: group_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_members
(
    id         character varying(64)                  NOT NULL,
    group_id   character varying                      NOT NULL,
    user_id    character varying(64)                  NOT NULL,
    role       character varying                      NOT NULL,
    status     integer                  DEFAULT 1     NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by character varying(64)                  NOT NULL,
    updated_by character varying(64)                  NOT NULL
);


--
-- Name: TABLE group_members; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON TABLE public.group_members IS '团队用户关系表';


--
-- Name: COLUMN group_members.role; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.group_members.role IS '角色, owner/maintainer/developer/viewer';


--
-- Name: identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.identities
(
    id              character varying(64)                  NOT NULL,
    user_id         character varying(64)                  NOT NULL,
    provider        character varying(256)                 NOT NULL,
    provider_id     character varying(256)                 NOT NULL,
    identity_data   jsonb,
    last_sign_in_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at      timestamp with time zone DEFAULT now() NOT NULL,
    updated_at      timestamp with time zone DEFAULT now() NOT NULL,
    created_by      character varying(64)                  NOT NULL,
    updated_by      character varying(64)                  NOT NULL
);


--
-- Name: TABLE identities; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON TABLE public.identities IS '用户第三方认证';


--
-- Name: COLUMN identities.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.identities.id IS '主键 (ULID)';


--
-- Name: COLUMN identities.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.identities.user_id IS '用户 UID (ULID)';


--
-- Name: COLUMN identities.provider; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.identities.provider IS '认证提供方';


--
-- Name: COLUMN identities.provider_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.identities.provider_id IS '第三方认证平台的用户唯一ID';


--
-- Name: COLUMN identities.identity_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.identities.identity_data IS '认证的元数据';


--
-- Name: COLUMN identities.last_sign_in_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.identities.last_sign_in_at IS '最近登录时间戳';


CREATE TABLE public.files
(
    id         character varying(128)                 NOT NULL,
    file       character varying(512)                 NOT NULL,
    bucket     character varying(255)                 NOT NULL,
    category   character varying                      NOT NULL,
    filetype   character varying                      NOT NULL,
    status     integer                  DEFAULT 1     NOT NULL,
    metadata   jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by character varying(64)                  NOT NULL,
    updated_by character varying(64)                  NOT NULL
);


--
-- Name: COLUMN files.file; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.files.file IS '文件名';


--
-- Name: COLUMN files.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON COLUMN public.files.category IS '分类';




--
-- Name: group_members group_members_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_members
    ADD CONSTRAINT group_members_pk PRIMARY KEY (id);

ALTER TABLE ONLY public.group_members
    ADD CONSTRAINT group_members_key UNIQUE (group_id, user_id);


--
-- Name: groups groups_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_name_key UNIQUE (name);


--
-- Name: groups groups_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pk PRIMARY KEY (id);


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_tokens
(
    id         character varying(255)                 NOT NULL,
    user_id    character varying(64)                  NOT NULL,
    token      character varying(255)                 NOT NULL,
    revoked    boolean                  DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE refresh_tokens; Type: COMMENT; Schema: public; Owner: -
--

COMMENT
ON TABLE public.refresh_tokens IS 'Refresh Token';


--
-- Name: refresh_tokens refresh_tokens_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pk PRIMARY KEY (id);

