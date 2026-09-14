CREATE TABLE IF NOT EXISTS sys_user (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'member',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_users_role (role),
  INDEX idx_users_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Users';

CREATE TABLE IF NOT EXISTS question (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  type VARCHAR(20) NOT NULL COMMENT 'single 单选 / multiple 多选 / judge 判断',
  content TEXT NOT NULL COMMENT '题干',
  options JSON NULL COMMENT '选项: [{"key":"A","text":"..."}]，判断题为空',
  answer JSON NOT NULL COMMENT '答案: 单选 "A" / 多选 ["A","C"] / 判断 true|false',
  analysis TEXT NULL COMMENT '答案解析',
  tags VARCHAR(255) NULL COMMENT '标签，逗号分隔',
  difficulty TINYINT NOT NULL DEFAULT 3 COMMENT '难度 1-5',
  default_score INT NOT NULL DEFAULT 5 COMMENT '默认分值',
  status VARCHAR(20) NOT NULL DEFAULT 'enabled' COMMENT 'enabled 启用 / disabled 停用',
  created_by BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_question_type (type),
  INDEX idx_question_status (status),
  INDEX idx_question_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目';

CREATE TABLE IF NOT EXISTS paper (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT 'draft 草稿 / published 已发布 / offline 已下线',
  total_score INT NOT NULL DEFAULT 0 COMMENT '总分（组卷时计算）',
  question_count INT NOT NULL DEFAULT 0 COMMENT '题目数量',
  created_by BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_paper_status (status),
  INDEX idx_paper_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='试卷';

CREATE TABLE IF NOT EXISTS paper_question (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  paper_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  seq INT NOT NULL COMMENT '题号，从 1 开始',
  score INT NOT NULL COMMENT '卷内分值',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_paper_question (paper_id, question_id),
  INDEX idx_paper_question_paper (paper_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='试卷题目关联';

CREATE TABLE IF NOT EXISTS exam_attempt (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  paper_id BIGINT NOT NULL,
  total_score INT NOT NULL DEFAULT 0 COMMENT '总得分',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
  INDEX idx_attempt_user (user_id),
  INDEX idx_attempt_paper (paper_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='答卷记录';

CREATE TABLE IF NOT EXISTS attempt_answer (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  attempt_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  user_answer JSON NULL COMMENT '用户答案，结构同 question.answer',
  is_correct TINYINT(1) NOT NULL DEFAULT 0,
  score INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_attempt_question (attempt_id, question_id),
  INDEX idx_attempt_answer_attempt (attempt_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='答卷作答明细';
