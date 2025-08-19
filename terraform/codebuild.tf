# Optional: CodeBuild project to build & push image to ECR
# Requires existing ECR repo (created manually or via separate resource)

variable "enable_codebuild" { type = bool default = true description = "Set false to skip creating CodeBuild project" }
variable "codebuild_service_role_arn" { type = string default = "" description = "Optional existing IAM role ARN for CodeBuild; if empty a role is created" }

locals {
  use_existing_codebuild_role = length(var.codebuild_service_role_arn) > 0
  create_codebuild            = var.enable_codebuild
}

# Create IAM role for CodeBuild if none supplied
data "aws_iam_policy_document" "codebuild_assume" {
  count = local.create_codebuild && !local.use_existing_codebuild_role ? 1 : 0
  statement {
    effect = "Allow"
    principals { type = "Service" identifiers = ["codebuild.amazonaws.com"] }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "codebuild" {
  count              = local.create_codebuild && !local.use_existing_codebuild_role ? 1 : 0
  name               = "${var.project}-codebuild-role"
  assume_role_policy = data.aws_iam_policy_document.codebuild_assume[0].json
}

data "aws_iam_policy_document" "codebuild_inline" {
  count = local.create_codebuild && !local.use_existing_codebuild_role ? 1 : 0
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup","logs:CreateLogStream","logs:PutLogEvents"
    ]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability","ecr:CompleteLayerUpload","ecr:UploadLayerPart","ecr:InitiateLayerUpload","ecr:PutImage","ecr:BatchGetImage",
      "ecr:DescribeRepositories","ecr:DescribeImages","ecr:GetDownloadUrlForLayer"
    ]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "codebuild_inline" {
  count  = local.create_codebuild && !local.use_existing_codebuild_role ? 1 : 0
  name   = "${var.project}-codebuild-inline"
  role   = aws_iam_role.codebuild[0].id
  policy = data.aws_iam_policy_document.codebuild_inline[0].json
}

resource "aws_codebuild_project" "image" {
  count         = local.create_codebuild ? 1 : 0
  name          = "${var.project}-image-build"
  description   = "Builds and pushes Docker image to ECR"
  service_role  = local.use_existing_codebuild_role ? var.codebuild_service_role_arn : aws_iam_role.codebuild[0].arn
  build_timeout = 30

  artifacts { type = "NO_ARTIFACTS" }
  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:7.0"
    type                        = "LINUX_CONTAINER"
    privileged_mode             = true
    image_pull_credentials_type = "CODEBUILD"
    environment_variable { name = "AWS_DEFAULT_REGION" value = var.region }
  }
  source {
    type      = "GITHUB"
    location  = "https://github.com/Hex-Zero/MaxwellGoSpine.git"
    buildspec = "codebuild-buildspec.yml"
  }

  logs_config { cloudwatch_logs { group_name = "/codebuild/${var.project}" stream_name = "image" } }

  tags = { Project = var.project }
}

output "codebuild_project_name" {
  value       = local.create_codebuild ? aws_codebuild_project.image[0].name : ""
  description = "Name of the CodeBuild project (if created)"
}
