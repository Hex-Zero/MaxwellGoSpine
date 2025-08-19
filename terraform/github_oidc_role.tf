# IAM role for GitHub Actions OIDC to push to ECR
# Create only if variable github_repo is set (format owner/name)

variable "github_repo" { type = string default = "" description = "GitHub repo (owner/name) to allow OIDC access" }

locals { create_github_oidc = length(var.github_repo) > 0 }

data "aws_caller_identity" "current" {}

# Create GitHub OIDC provider if not already present (safe if previously created elsewhere skip toggle externally)
data "aws_iam_openid_connect_provider" "github" {
  count = 0 # placeholder when already exists; not discoverable reliably without errors in pure Terraform w/out data source
  arn   = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"
}

resource "aws_iam_openid_connect_provider" "github" {
  count           = local.create_github_oidc ? 1 : 0
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
  tags            = { Name = "github-actions" }
}

data "aws_iam_policy_document" "github_oidc_assume" {
  count = local.create_github_oidc ? 1 : 0
  statement {
    effect = "Allow"
    principals {
      type        = "Federated"
      identifiers = [try(aws_iam_openid_connect_provider.github[0].arn, "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com")]
    }
    actions = ["sts:AssumeRoleWithWebIdentity"]
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repo}:ref:refs/heads/main", "repo:${var.github_repo}:pull_request"]
    }
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "github_oidc" {
  count              = local.create_github_oidc ? 1 : 0
  name               = "${var.project}-github-oidc"
  assume_role_policy = data.aws_iam_policy_document.github_oidc_assume[0].json
}

# Policy: Allow ECR push & STS read (narrow set)
data "aws_iam_policy_document" "github_oidc_policy" {
  count = local.create_github_oidc ? 1 : 0
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:CompleteLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:InitiateLayerUpload",
      "ecr:PutImage",
      "ecr:BatchGetImage",
      "ecr:DescribeRepositories",
      "ecr:DescribeImages",
      "ecr:GetDownloadUrlForLayer"
    ]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "github_oidc" {
  count       = local.create_github_oidc ? 1 : 0
  name        = "${var.project}-github-oidc-policy"
  description = "Permissions for GitHub Actions to push to ECR"
  policy      = data.aws_iam_policy_document.github_oidc_policy[0].json
}

resource "aws_iam_role_policy_attachment" "github_oidc_attach" {
  count      = local.create_github_oidc ? 1 : 0
  role       = aws_iam_role.github_oidc[0].name
  policy_arn = aws_iam_policy.github_oidc[0].arn
}

output "github_actions_oidc_role_arn" {
  value       = try(aws_iam_role.github_oidc[0].arn, "")
  description = "Role ARN to place in GitHub secret AWS_GITHUB_OIDC_ROLE_ARN"
}
