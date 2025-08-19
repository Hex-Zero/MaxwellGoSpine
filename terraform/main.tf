###############################
# Terraform ECS Fargate Stack
###############################

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
  profile = var.profile
}

################
# Input Variables
################

variable "region" { type = string default = "us-east-1" }
variable "profile" { type = string default = "maxwell" }
variable "project" { type = string default = "maxwell" }
variable "environment" { type = string default = "prod" }
variable "api_port" { type = number default = 8080 }
variable "desired_count" { type = number default = 2 }
variable "max_count" { type = number default = 6 }
variable "cpu" { type = number default = 512 }      # 0.5 vCPU
variable "memory" { type = number default = 1024 }  # MiB
variable "api_image" { type = string description = "ECR image URI (e.g. ACCOUNT.dkr.ecr.region.amazonaws.com/maxwell-api:tag)" }
variable "db_dsn" { type = string sensitive = true }
variable "api_keys" { type = string sensitive = true }

#############
# Networking
#############

resource "aws_vpc" "main" {
  cidr_block           = "10.42.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = { Name = "${var.project}-vpc" }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id
  tags = { Name = "${var.project}-igw" }
}

resource "aws_subnet" "public_a" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.42.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true
  tags = { Name = "${var.project}-public-a" }
}

resource "aws_subnet" "public_b" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.42.2.0/24"
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = true
  tags = { Name = "${var.project}-public-b" }
}

data "aws_availability_zones" "available" { state = "available" }

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  tags = { Name = "${var.project}-public-rt" }
}

resource "aws_route" "default_inet" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

resource "aws_route_table_association" "a" { subnet_id = aws_subnet.public_a.id route_table_id = aws_route_table.public.id }
resource "aws_route_table_association" "b" { subnet_id = aws_subnet.public_b.id route_table_id = aws_route_table.public.id }

########################
# Security Group & LB
########################

resource "aws_security_group" "lb" {
  name        = "${var.project}-lb-sg"
  description = "ALB SG"
  vpc_id      = aws_vpc.main.id
  ingress { from_port = 80 to_port = 80 protocol = "tcp" cidr_blocks = ["0.0.0.0/0"] }
  egress  { from_port = 0  to_port = 0  protocol = "-1" cidr_blocks = ["0.0.0.0/0"] }
  tags = { Name = "${var.project}-lb-sg" }
}

resource "aws_security_group" "ecs" {
  name        = "${var.project}-ecs-sg"
  description = "ECS tasks"
  vpc_id      = aws_vpc.main.id
  ingress { from_port = var.api_port to_port = var.api_port protocol = "tcp" security_groups = [aws_security_group.lb.id] }
  egress  { from_port = 0 to_port = 0 protocol = "-1" cidr_blocks = ["0.0.0.0/0"] }
  tags = { Name = "${var.project}-ecs-sg" }
}

resource "aws_lb" "app" {
  name               = "${var.project}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.lb.id]
  subnets            = [aws_subnet.public_a.id, aws_subnet.public_b.id]
  tags = { Name = "${var.project}-alb" }
}

resource "aws_lb_target_group" "api" {
  name        = "${var.project}-tg"
  port        = var.api_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"
  health_check {
    path                = "/healthz"
    matcher             = "200"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.app.arn
  port              = 80
  protocol          = "HTTP"
  default_action { type = "forward" target_group_arn = aws_lb_target_group.api.arn }
}

########################
# CloudWatch Logs
########################
resource "aws_cloudwatch_log_group" "api" { name = "/ecs/${var.project}" retention_in_days = 14 }

########################
# IAM Roles
########################

data "aws_iam_policy_document" "ecs_task_assume" {
  statement {
    effect = "Allow"
    principals { type = "Service" identifiers = ["ecs-tasks.amazonaws.com"] }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "execution" {
  name               = "${var.project}-exec-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_assume.json
}

resource "aws_iam_role_policy_attachment" "exec_default" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# (Optional) add SSM read if using ssm parameters for secrets.
resource "aws_iam_role_policy_attachment" "exec_ssm" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess"
}

########################
# SSM Parameters (secrets)
########################
resource "aws_ssm_parameter" "api_keys" {
  name  = "/${var.project}/API_KEYS"
  type  = "SecureString"
  value = var.api_keys
}
resource "aws_ssm_parameter" "db_dsn" {
  name  = "/${var.project}/DB_DSN"
  type  = "SecureString"
  value = var.db_dsn
}

########################
# ECS Cluster & Task
########################
resource "aws_ecs_cluster" "main" { name = "${var.project}-cluster" }

resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project}-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = tostring(var.cpu)
  memory                   = tostring(var.memory)
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.execution.arn

  container_definitions = jsonencode([
    {
      name  = "api"
      image = var.api_image
      essential = true
      portMappings = [{ containerPort = var.api_port, protocol = "tcp" }]
      environment = [
        { name = "ENV", value = var.environment },
        { name = "READ_TIMEOUT", value = "10s" },
        { name = "WRITE_TIMEOUT", value = "15s" }
      ]
      secrets = [
        { name = "API_KEYS", valueFrom = aws_ssm_parameter.api_keys.arn },
        { name = "DB_DSN",   valueFrom = aws_ssm_parameter.db_dsn.arn }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.api.name
          awslogs-region        = var.region
          awslogs-stream-prefix = "api"
        }
      }
      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:${var.api_port}/healthz || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 10
      }
    }
  ])
}

resource "aws_ecs_service" "api" {
  name            = "${var.project}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200

  network_configuration {
    subnets         = [aws_subnet.public_a.id, aws_subnet.public_b.id]
    security_groups = [aws_security_group.ecs.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = var.api_port
  }

  lifecycle { ignore_changes = [desired_count] }
}

########################
# Application Auto Scaling
########################
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = var.max_count
  min_capacity       = var.desired_count
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.api.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "cpu_scale" {
  name               = "${var.project}-cpu-target"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification { predefined_metric_type = "ECSServiceAverageCPUUtilization" }
    target_value       = 60
    scale_in_cooldown  = 60
    scale_out_cooldown = 60
  }
}

########################
# Outputs
########################
output "alb_dns_name" { value = aws_lb.app.dns_name }
output "api_url"      { value = "http://${aws_lb.app.dns_name}" }
