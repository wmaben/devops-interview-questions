# Top 15 Terraform Interview Questions for a Senior DevOps Engineer (5+ Years)

---

## 🏗️ CATEGORY 1: Terraform Core Concepts & Internals

---

### 1. Explain Terraform's Architecture and How It Works Internally

**Expected Answer:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Terraform Architecture                         │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   Terraform Core                            │   │
│  │                                                             │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │   │
│  │  │  HCL Parser  │  │  Graph       │  │  State Manager   │  │   │
│  │  │  (Config     │  │  Builder     │  │  (local/remote)  │  │   │
│  │  │   Loading)   │  │  (DAG)       │  │                  │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────────────┘  │   │
│  │                                                             │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │   │
│  │  │  Plan Engine │  │  Apply       │  │  Diff Engine     │  │   │
│  │  │  (desired vs │  │  Engine      │  │  (state vs       │  │   │
│  │  │  actual)     │  │              │  │   config)        │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────────────┘  │   │
│  └───────────────────────────┬─────────────────────────────────┘   │
│                              │ Plugin Protocol (gRPC)              │
│  ┌───────────────────────────▼─────────────────────────────────┐   │
│  │                   Provider Plugins                          │   │
│  │                                                             │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │   │
│  │  │  AWS     │  │  Azure   │  │  GCP     │  │  K8s     │   │   │
│  │  │ Provider │  │ Provider │  │ Provider │  │ Provider │   │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                     │
│  ┌───────────────────────────▼─────────────────────────────────┐   │
│  │              Cloud Provider APIs / Services                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

**Terraform Execution Flow:**

```
terraform init
    │
    ├── Download providers (registry.terraform.io)
    ├── Initialize backend (S3, GCS, Terraform Cloud)
    └── Install modules

terraform plan
    │
    ├── Load configuration files (.tf)
    ├── Load & lock state file
    ├── Build dependency graph (DAG)
    ├── Call provider Read APIs (refresh current state)
    ├── Diff: desired config vs actual state
    └── Output execution plan

terraform apply
    │
    ├── Walk the dependency graph
    ├── Execute operations in parallel (where no dependency)
    ├── Call provider CRUD APIs
    ├── Update state file after each resource
    └── Output results
```

**Dependency Graph (DAG):**
```hcl
# Terraform builds a Directed Acyclic Graph automatically
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "public" {
  vpc_id = aws_vpc.main.id    # Implicit dependency — subnet depends on VPC
  cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "app" {
  vpc_id = aws_vpc.main.id    # Also depends on VPC
}

resource "aws_instance" "app" {
  subnet_id         = aws_subnet.public.id          # Depends on subnet
  security_group_ids = [aws_security_group.app.id]  # Depends on SG

  # Explicit dependency (no reference in config)
  depends_on = [aws_internet_gateway.main]
}
```

```bash
# Visualize the dependency graph
terraform graph | dot -Tsvg > graph.svg

# Visualize with specific plan
terraform graph -type=plan | dot -Tpng > plan-graph.png
```

---

### 2. Explain Terraform State — Why It Exists, Remote Backends, and State Locking

**Expected Answer:**

**Why State Exists:**
```
State is Terraform's source of truth about real-world infrastructure.
It maps your configuration to actual cloud resources.

Without state:
❌ Terraform can't know what already exists
❌ Can't detect drift between config and reality
❌ Can't determine what needs to create/update/delete
❌ Can't track resource metadata (IDs, ARNs, IPs)
```

**State File Example:**
```json
{
  "version": 4,
  "terraform_version": "1.6.0",
  "serial": 42,
  "lineage": "uuid-identifies-state-uniquely",
  "outputs": {
    "vpc_id": {
      "value": "vpc-0a1b2c3d4e5f",
      "type": "string"
    }
  },
  "resources": [
    {
      "mode": "managed",
      "type": "aws_vpc",
      "name": "main",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "vpc-0a1b2c3d4e5f",
            "cidr_block": "10.0.0.0/16",
            "arn": "arn:aws:ec2:us-east-1:123456789:vpc/vpc-0a1b2c3d4e5f",
            "tags": {
              "Name": "main-vpc",
              "Environment": "production"
            }
          }
        }
      ]
    }
  ]
}
```

**Remote Backend with S3 + DynamoDB Locking:**
```hcl
# backend.tf — Production-grade S3 backend
terraform {
  backend "s3" {
    bucket         = "company-terraform-state"
    key            = "environments/production/us-east-1/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true                       # Server-side encryption
    kms_key_id     = "arn:aws:kms:us-east-1:123:key/xxx"

    # State locking via DynamoDB
    dynamodb_table = "terraform-state-locks"

    # Access logging
    acl            = "private"

    # Versioning (enable on bucket for state history)
    # S3 bucket versioning must be enabled separately
  }

  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

```hcl
# Bootstrap: Create the S3 bucket and DynamoDB table
resource "aws_s3_bucket" "terraform_state" {
  bucket = "company-terraform-state"

  lifecycle {
    prevent_destroy = true    # Never accidentally delete state bucket
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled"        # Keep all state versions
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.terraform.arn
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket                  = aws_s3_bucket.terraform_state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-state-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "Terraform State Lock Table"
  }
}
```

**State Management Commands:**
```bash
# List all resources in state
terraform state list

# Show a specific resource's state
terraform state show aws_instance.app

# Move resource in state (rename without recreate)
terraform state mv aws_instance.app aws_instance.web_server

# Remove a resource from state (stop managing without destroying)
terraform state rm aws_instance.old_server

# Import existing resource into state
terraform import aws_instance.imported i-0a1b2c3d4e5f

# Pull remote state to local
terraform state pull > backup-state.json

# Push local state to remote (dangerous!)
terraform state push backup-state.json

# Force unlock state (if lock is stuck)
terraform force-unlock <LOCK_ID>

# Refresh state against real infrastructure
terraform apply -refresh-only

# Show state differences (drift detection)
terraform plan -refresh-only
```

---

### 3. How Does Terraform Handle Provider Authentication and Multi-Region/Multi-Account Deployments?

**Expected Answer:**

**Provider Configuration:**
```hcl
# Single provider with alias for multi-region
provider "aws" {
  region = "us-east-1"
  # Auth priority: env vars → shared credentials → IAM role → instance profile
}

provider "aws" {
  alias  = "us-west-2"
  region = "us-west-2"
}

provider "aws" {
  alias  = "eu-west-1"
  region = "eu-west-1"
}

# Use aliased provider in resource
resource "aws_instance" "west_app" {
  provider      = aws.us-west-2
  ami           = "ami-12345678"
  instance_type = "t3.micro"
}
```

**Multi-Account with AssumeRole:**
```hcl
# Master account provider
provider "aws" {
  region = "us-east-1"
  alias  = "master"
}

# Production account — assume role
provider "aws" {
  region = "us-east-1"
  alias  = "production"

  assume_role {
    role_arn     = "arn:aws:iam::111111111111:role/TerraformDeployRole"
    session_name = "terraform-production"
    duration     = "1h"
    external_id  = var.external_id    # Extra security
  }

  default_tags {
    tags = {
      ManagedBy   = "terraform"
      Environment = "production"
      Account     = "111111111111"
    }
  }
}

# Staging account
provider "aws" {
  region = "us-east-1"
  alias  = "staging"

  assume_role {
    role_arn     = "arn:aws:iam::222222222222:role/TerraformDeployRole"
    session_name = "terraform-staging"
  }
}

# Use specific account's provider
resource "aws_vpc" "production_vpc" {
  provider   = aws.production
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "staging_vpc" {
  provider   = aws.staging
  cidr_block = "10.1.0.0/16"
}
```

**Authentication Methods (Priority Order):**
```bash
# Method 1: Environment variables (CI/CD pipelines)
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="xxx"
export AWS_SESSION_TOKEN="yyy"   # For temporary credentials

# Method 2: AWS Profile
provider "aws" {
  profile = "production"
  region  = "us-east-1"
}

# Method 3: OIDC / IRSA (Kubernetes)
# Method 4: EC2 Instance Profile / ECS Task Role (no creds needed)
# Method 5: Terraform Cloud dynamic credentials (recommended for TFC)

# Best practice for CI/CD: OIDC Federation — no static credentials!
# GitHub Actions OIDC with AWS
provider "aws" {
  region = "us-east-1"
  # Credentials automatically from OIDC token
}
```

---

## 🔧 CATEGORY 2: Modules & Code Organization

---

### 4. How Do You Design and Structure Production-Grade Terraform Modules?

**Expected Answer:**

**Repository Structure:**
```
terraform-infrastructure/
├── modules/                     # Reusable modules
│   ├── networking/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── versions.tf
│   │   ├── README.md
│   │   └── examples/
│   │       └── complete/
│   │           ├── main.tf
│   │           └── outputs.tf
│   ├── compute/
│   │   ├── ec2/
│   │   └── eks/
│   ├── database/
│   │   ├── rds/
│   │   └── dynamodb/
│   └── security/
│       ├── iam/
│       └── kms/
│
├── environments/                # Environment-specific configs
│   ├── production/
│   │   ├── us-east-1/
│   │   │   ├── main.tf
│   │   │   ├── variables.tf
│   │   │   ├── outputs.tf
│   │   │   ├── backend.tf
│   │   │   └── terraform.tfvars
│   │   └── eu-west-1/
│   ├── staging/
│   └── development/
│
└── global/                      # Global resources (IAM, Route53)
    ├── iam/
    └── dns/
```

**Well-Designed Module — networking/main.tf:**
```hcl
# modules/networking/main.tf

locals {
  # Compute derived values once
  name_prefix = "${var.environment}-${var.project}"

  # Auto-generate subnet CIDRs
  public_subnets  = [for i, az in var.availability_zones :
    cidrsubnet(var.vpc_cidr, 4, i)]
  private_subnets = [for i, az in var.availability_zones :
    cidrsubnet(var.vpc_cidr, 4, i + 10)]
  database_subnets = [for i, az in var.availability_zones :
    cidrsubnet(var.vpc_cidr, 4, i + 20)]

  common_tags = merge(var.tags, {
    Module      = "networking"
    Environment = var.environment
    ManagedBy   = "terraform"
  })
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-vpc"
    "kubernetes.io/cluster/${local.name_prefix}" = "shared"  # EKS tag
  })
}

resource "aws_subnet" "public" {
  count = length(var.availability_zones)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = local.public_subnets[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-public-${var.availability_zones[count.index]}"
    Tier = "public"
    "kubernetes.io/role/elb" = "1"   # EKS public LB tag
  })
}

resource "aws_subnet" "private" {
  count = length(var.availability_zones)

  vpc_id            = aws_vpc.main.id
  cidr_block        = local.private_subnets[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-private-${var.availability_zones[count.index]}"
    Tier = "private"
    "kubernetes.io/role/internal-elb" = "1"
  })
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-igw"
  })
}

resource "aws_eip" "nat" {
  count  = var.enable_nat_gateway ? length(var.availability_zones) : 0
  domain = "vpc"

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-eip-${count.index + 1}"
  })

  depends_on = [aws_internet_gateway.main]
}

resource "aws_nat_gateway" "main" {
  count = var.enable_nat_gateway ? length(var.availability_zones) : 0

  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-nat-${var.availability_zones[count.index]}"
  })
}
```

**Module Variables — variables.tf:**
```hcl
# modules/networking/variables.tf

variable "environment" {
  description = "Environment name (production, staging, development)"
  type        = string

  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "Environment must be one of: production, staging, development."
  }
}

variable "project" {
  description = "Project name used in resource naming"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]{2,20}$", var.project))
    error_message = "Project name must be 2-20 lowercase alphanumeric characters or hyphens."
  }
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "Must be a valid CIDR block."
  }
}

variable "availability_zones" {
  description = "List of availability zones for subnet creation"
  type        = list(string)

  validation {
    condition     = length(var.availability_zones) >= 2
    error_message = "At least 2 availability zones required for high availability."
  }
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway for private subnet internet access"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
```

**Module Outputs — outputs.tf:**
```hcl
# modules/networking/outputs.tf

output "vpc_id" {
  description = "ID of the created VPC"
  value       = aws_vpc.main.id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = aws_subnet.private[*].id
}

output "nat_gateway_ids" {
  description = "List of NAT Gateway IDs"
  value       = aws_nat_gateway.main[*].id
}

# Useful for EKS
output "subnet_ids_by_az" {
  description = "Map of AZ to private subnet ID"
  value = {
    for i, az in var.availability_zones :
    az => aws_subnet.private[i].id
  }
}
```

**Consuming the Module:**
```hcl
# environments/production/us-east-1/main.tf

module "networking" {
  source = "../../../modules/networking"
  # Or from Terraform Registry
  # source  = "terraform-aws-modules/vpc/aws"
  # version = "~> 5.0"

  environment        = "production"
  project            = "myapp"
  vpc_cidr           = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  enable_nat_gateway = true

  tags = {
    Team       = "platform"
    CostCenter = "infrastructure"
    Owner      = "platform-team@company.com"
  }
}

module "eks" {
  source = "../../../modules/compute/eks"

  # Pass networking outputs to EKS module
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  public_subnet_ids  = module.networking.public_subnet_ids
}
```

---

### 5. Explain Terraform Workspaces vs Directory-Based Environment Separation

**Expected Answer:**

**Terraform Workspaces:**
```bash
# Create and switch workspaces
terraform workspace new production
terraform workspace new staging
terraform workspace new development

# List workspaces
terraform workspace list
# * default
#   development
#   production
#   staging

# Switch workspace
terraform workspace select production

# Show current workspace
terraform workspace show
```

```hcl
# Using workspace in configuration
locals {
  # Workspace-based configuration map
  workspace_config = {
    production = {
      instance_type  = "m5.2xlarge"
      min_capacity   = 3
      max_capacity   = 50
      rds_class      = "db.r5.2xlarge"
      multi_az       = true
      deletion_protection = true
    }
    staging = {
      instance_type  = "t3.large"
      min_capacity   = 1
      max_capacity   = 10
      rds_class      = "db.t3.medium"
      multi_az       = false
      deletion_protection = false
    }
    development = {
      instance_type  = "t3.small"
      min_capacity   = 1
      max_capacity   = 3
      rds_class      = "db.t3.micro"
      multi_az       = false
      deletion_protection = false
    }
  }

  config = local.workspace_config[terraform.workspace]
}

resource "aws_instance" "app" {
  instance_type = local.config.instance_type
  # ...
}
```

**Why Directory-Based is Better for Production:**

```
❌ Workspaces Problems:
  - Same backend, accidental apply to wrong environment
  - Shared provider versions across environments
  - No blast radius isolation
  - Single state file location
  - Cannot have different backend configs

✅ Directory-Based Advantages:
  - Complete isolation between environments
  - Different backends per environment
  - Different provider versions possible
  - Independent state files
  - Full blast radius isolation
  - Easier CI/CD pipeline targeting
```

```
environments/
├── production/
│   ├── backend.tf          # → s3://state-bucket/prod/terraform.tfstate
│   └── terraform.tfvars    # Production-specific values
├── staging/
│   ├── backend.tf          # → s3://state-bucket/staging/terraform.tfstate
│   └── terraform.tfvars    # Staging-specific values
└── development/
    ├── backend.tf          # → s3://state-bucket/dev/terraform.tfstate
    └── terraform.tfvars    # Development-specific values
```

---

## 🔄 CATEGORY 3: Advanced HCL Features

---

### 6. Explain Advanced HCL Features — Dynamic Blocks, for_each, Conditional Expressions, and Functions

**Expected Answer:**

**for_each vs count:**
```hcl
# ❌ count — fragile, index-based
resource "aws_subnet" "public" {
  count      = 3
  cidr_block = "10.0.${count.index}.0/24"
  # Deleting index 1 causes index 2 to become index 1 → RECREATION
}

# ✅ for_each — stable, key-based
resource "aws_subnet" "public" {
  for_each = {
    "us-east-1a" = "10.0.1.0/24"
    "us-east-1b" = "10.0.2.0/24"
    "us-east-1c" = "10.0.3.0/24"
  }

  availability_zone = each.key
  cidr_block        = each.value
  # Deleting one AZ only removes that specific subnet
}

# for_each with toset (list → set)
resource "aws_security_group_rule" "ingress" {
  for_each = toset(var.allowed_ports)

  type              = "ingress"
  from_port         = each.value
  to_port           = each.value
  protocol          = "tcp"
  security_group_id = aws_security_group.main.id
  cidr_blocks       = ["0.0.0.0/0"]
}

# for_each with complex objects
variable "microservices" {
  default = {
    api = {
      port     = 8080
      replicas = 3
      image    = "api:v1.0"
    }
    worker = {
      port     = 8081
      replicas = 5
      image    = "worker:v1.0"
    }
  }
}

resource "kubernetes_deployment" "services" {
  for_each = var.microservices

  metadata {
    name = each.key
  }

  spec {
    replicas = each.value.replicas
    # ...
  }
}
```

**Dynamic Blocks:**
```hcl
# ❌ Without dynamic blocks — repetitive
resource "aws_security_group" "app" {
  name = "app-sg"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }
}

# ✅ With dynamic blocks — flexible and DRY
variable "ingress_rules" {
  default = [
    { port = 80,   cidr = "0.0.0.0/0",   description = "HTTP" },
    { port = 443,  cidr = "0.0.0.0/0",   description = "HTTPS" },
    { port = 8080, cidr = "10.0.0.0/8",  description = "Internal API" },
  ]
}

resource "aws_security_group" "app" {
  name = "app-sg"

  dynamic "ingress" {
    for_each = var.ingress_rules
    content {
      from_port   = ingress.value.port
      to_port     = ingress.value.port
      protocol    = "tcp"
      cidr_blocks = [ingress.value.cidr]
      description = ingress.value.description
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

**Advanced Expressions & Functions:**
```hcl
locals {
  # Conditional expression
  instance_type = var.environment == "production" ? "m5.2xlarge" : "t3.medium"

  # Complex conditionals with try()
  db_password = try(data.aws_secretsmanager_secret_version.db.secret_string, "default-dev-pass")

  # String functions
  name_upper     = upper(var.project)
  name_sanitized = replace(var.project, "-", "_")
  name_truncated = substr(var.project, 0, 20)

  # Collection functions
  unique_azs   = distinct(var.availability_zones)
  sorted_azs   = sort(var.availability_zones)
  az_count     = length(var.availability_zones)
  first_az     = element(var.availability_zones, 0)

  # Map operations
  merged_tags = merge(var.common_tags, var.extra_tags, {
    Timestamp = timestamp()
  })

  # List to map transformation
  subnet_map = { for subnet in aws_subnet.private : subnet.availability_zone => subnet.id }

  # Flatten nested lists
  all_cidrs = flatten([
    [for subnet in aws_subnet.public : subnet.cidr_block],
    [for subnet in aws_subnet.private : subnet.cidr_block]
  ])

  # Filtering with for + if
  production_instances = {
    for k, v in var.instances : k => v
    if v.environment == "production"
  }

  # one() — extract single value or null
  primary_subnet = one([for s in aws_subnet.public : s.id if s.tags["Primary"] == "true"])

  # coalesce — first non-null value
  region = coalesce(var.region, data.aws_region.current.name, "us-east-1")

  # coalescelist — first non-empty list
  subnet_ids = coalescelist(var.custom_subnets, aws_subnet.private[*].id)

  # formatlist
  instance_names = formatlist("web-%s-%03d", var.environment, range(var.instance_count))

  # jsonencode / yamlencode
  policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:PutObject"]
        Resource = "${aws_s3_bucket.main.arn}/*"
      }
    ]
  })
}
```

---

### 7. Explain Terraform Meta-Arguments — lifecycle, depends_on, provider, and count/for_each

**Expected Answer:**

```hcl
resource "aws_instance" "web" {
  ami           = var.ami_id
  instance_type = var.instance_type

  # ─── lifecycle ────────────────────────────────────────────
  lifecycle {
    # Don't destroy before creating replacement (zero-downtime)
    create_before_destroy = true

    # Never allow Terraform to destroy this resource
    prevent_destroy = true

    # Ignore changes to these attributes (avoid drift issues)
    ignore_changes = [
      ami,              # Don't replace if AMI updates
      user_data,        # Don't replace on user_data change
      tags["LastDeploy"] # Ignore auto-updated tags
    ]

    # Custom pre-conditions (Terraform 1.2+)
    precondition {
      condition     = var.instance_type != "t2.micro"
      error_message = "t2.micro is not allowed in production."
    }

    # Custom post-conditions — verify after creation
    postcondition {
      condition     = self.public_ip != ""
      error_message = "Instance must have a public IP address."
    }

    # Replace resource if condition changes (Terraform 1.2+)
    replace_triggered_by = [
      aws_launch_template.main.id   # Replace instance if launch template changes
    ]
  }
}

# ─── depends_on ────────────────────────────────────────────
resource "aws_eks_node_group" "workers" {
  cluster_name = aws_eks_cluster.main.name

  # Explicit dependency — ensure IAM policies attached before node group
  depends_on = [
    aws_iam_role_policy_attachment.worker_node_policy,
    aws_iam_role_policy_attachment.cni_policy,
    aws_iam_role_policy_attachment.registry_policy,
  ]
}

# ─── moved block (Terraform 1.1+) ─────────────────────────
# Rename resource without destroying
moved {
  from = aws_instance.app
  to   = aws_instance.web_server
}

# Move resource into module
moved {
  from = aws_security_group.app
  to   = module.security.aws_security_group.app
}
```

---

## 🔒 CATEGORY 4: Security & Secrets

---

### 8. How Do You Manage Secrets and Sensitive Data in Terraform Securely?

**Expected Answer:**

```hcl
# ─── Mark outputs as sensitive ──────────────────────────────
output "database_password" {
  value     = aws_db_instance.main.password
  sensitive = true    # Will show as (sensitive value) in output
}

# ─── Mark variables as sensitive ────────────────────────────
variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true    # Never logged/displayed
}

# ─── Sensitive in locals ─────────────────────────────────────
locals {
  connection_string = sensitive(
    "postgresql://${var.db_user}:${var.db_pass}@${aws_db_instance.main.endpoint}/mydb"
  )
}
```

**Fetching Secrets from AWS Secrets Manager:**
```hcl
# Fetch existing secret — never hardcode credentials
data "aws_secretsmanager_secret" "db_credentials" {
  name = "production/database/credentials"
}

data "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = data.aws_secretsmanager_secret.db_credentials.id
}

locals {
  db_creds = jsondecode(
    data.aws_secretsmanager_secret_version.db_credentials.secret_string
  )
}

resource "aws_db_instance" "main" {
  identifier     = "prod-database"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.r5.2xlarge"

  # Use secret from AWS Secrets Manager
  username = local.db_creds.username
  password = local.db_creds.password

  # Let AWS manage password rotation
  manage_master_user_password = true
  master_user_secret_kms_key_id = aws_kms_key.rds.arn
}
```

**HashiCorp Vault Integration:**
```hcl
provider "vault" {
  address = "https://vault.company.com"
  # Auth via AppRole or OIDC
}

data "vault_generic_secret" "database" {
  path = "secret/production/database"
}

resource "aws_db_instance" "main" {
  password = data.vault_generic_secret.database.data["password"]
}
```

**Sensitive Files — Never Commit:**
```gitignore
# .gitignore for Terraform
*.tfvars              # May contain secrets
*.tfvars.json
terraform.tfstate     # Contains ALL resource attributes
terraform.tfstate.backup
.terraform/           # Provider binaries
.terraform.lock.hcl   # OK to commit (provider hash lock)
*.tfplan              # May contain sensitive data
crash.log
override.tf
override.tf.json
*_override.tf
```

**Safe tfvars pattern:**
```hcl
# terraform.tfvars.example — commit this template
db_password     = "REPLACE_WITH_ACTUAL"
api_key         = "REPLACE_WITH_ACTUAL"

# terraform.tfvars — NEVER commit
db_password     = "actualS3cr3tP@ss"
api_key         = "sk-actual-api-key"
```

---

## 🔄 CATEGORY 5: CI/CD & GitOps

---

### 9. How Do You Implement a Production-Grade Terraform CI/CD Pipeline?

**Expected Answer:**

**Terraform GitOps Flow:**
```
┌─────────────────────────────────────────────────────────────┐
│                  Terraform CI/CD Pipeline                   │
│                                                             │
│  PR Created       CI Checks          PR Approved           │
│  ┌──────────┐    ┌──────────────┐    ┌───────────────────┐ │
│  │  Git     │───►│ fmt + lint   │───►│  Manual Approval  │ │
│  │  Push    │    │ validate     │    │  (production)     │ │
│  └──────────┘    │ plan         │    └────────┬──────────┘ │
│                  │ cost estimate│             │            │
│                  │ security scan│    ┌────────▼──────────┐ │
│                  └──────────────┘    │   terraform apply │ │
│                                      │   on merge to main│ │
│                                      └───────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**GitHub Actions Pipeline:**
```yaml
name: Terraform CI/CD

on:
  pull_request:
    paths:
      - 'environments/**'
      - 'modules/**'
  push:
    branches:
      - main
    paths:
      - 'environments/**'
      - 'modules/**'

env:
  TF_VERSION: "1.6.0"
  TF_WORKING_DIR: environments/production/us-east-1

permissions:
  id-token: write         # OIDC token for AWS
  contents: read
  pull-requests: write    # Comment plan on PR

jobs:
  terraform-check:
    name: "Terraform Checks"
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Configure AWS Credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789:role/GitHubActionsRole
          aws-region: us-east-1

      - name: Terraform Format Check
        id: fmt
        run: terraform fmt -check -recursive
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Terraform Init
        id: init
        run: terraform init -backend=true
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Terraform Validate
        id: validate
        run: terraform validate -no-color
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Run tfsec (Security Scanning)
        uses: aquasecurity/tfsec-action@v1.0.0
        with:
          working_directory: ${{ env.TF_WORKING_DIR }}
          soft_fail: false
          github_token: ${{ secrets.GITHUB_TOKEN }}

      - name: Run tflint
        uses: terraform-linters/setup-tflint@v4
      - run: |
          tflint --init
          tflint --recursive --format compact
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Terraform Plan
        id: plan
        run: |
          terraform plan \
            -no-color \
            -input=false \
            -out=tfplan \
            2>&1 | tee plan_output.txt

          # Save exit code
          echo "PLAN_EXIT_CODE=${PIPESTATUS[0]}" >> $GITHUB_ENV
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Infracost — Cost Estimation
        uses: infracost/actions/setup@v2
        with:
          api-key: ${{ secrets.INFRACOST_API_KEY }}
      - run: |
          infracost breakdown \
            --path=. \
            --format=json \
            --out-file=/tmp/infracost.json

          infracost comment github \
            --path=/tmp/infracost.json \
            --github-token=${{ github.token }} \
            --pull-request=${{ github.event.pull_request.number }} \
            --repo=${{ github.repository }} \
            --behavior=update
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Comment Plan on PR
        uses: actions/github-script@v7
        if: github.event_name == 'pull_request'
        env:
          PLAN: ${{ steps.plan.outputs.stdout }}
        with:
          script: |
            const output = `#### Terraform Plan 📖
            #### Format 🖌 \`${{ steps.fmt.outcome }}\`
            #### Validate ⚙️ \`${{ steps.validate.outcome }}\`
            
            <details><summary>Show Plan</summary>
            
            \`\`\`terraform\n
            ${process.env.PLAN}
            \`\`\`
            
            </details>
            
            *Pushed by: @${{ github.actor }}*`;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            })

  terraform-apply:
    name: "Terraform Apply"
    runs-on: ubuntu-latest
    needs: [terraform-check]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment: production    # Requires manual approval in GitHub

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Configure AWS Credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789:role/GitHubActionsRole
          aws-region: us-east-1

      - name: Terraform Init
        run: terraform init
        working-directory: ${{ env.TF_WORKING_DIR }}

      - name: Terraform Apply
        run: |
          terraform apply \
            -auto-approve \
            -input=false \
            -parallelism=20
        working-directory: ${{ env.TF_WORKING_DIR }}
```

---

## 🏛️ CATEGORY 6: State Management & Refactoring

---

### 10. How Do You Safely Refactor Terraform Code Without Causing Infrastructure Downtime?

**Expected Answer:**

```hcl
# ─── Scenario 1: Rename a resource ──────────────────────────

# Before:
resource "aws_s3_bucket" "data" {
  bucket = "my-data-bucket"
}

# After rename (using moved block — NO destruction):
resource "aws_s3_bucket" "application_data" {
  bucket = "my-data-bucket"
}

moved {
  from = aws_s3_bucket.data
  to   = aws_s3_bucket.application_data
}

# ─── Scenario 2: Extract to module ──────────────────────────

# Before (resource in root):
resource "aws_security_group" "app" {
  name   = "app-sg"
  vpc_id = var.vpc_id
}

# After (moved into module):
module "security" {
  source = "./modules/security"
  vpc_id = var.vpc_id
}

moved {
  from = aws_security_group.app
  to   = module.security.aws_security_group.app
}

# ─── Scenario 3: count → for_each migration ─────────────────

# Before (count):
resource "aws_subnet" "private" {
  count      = 3
  cidr_block = "10.0.${count.index}.0/24"
}

# After (for_each): Requires state manipulation
resource "aws_subnet" "private" {
  for_each   = toset(["us-east-1a", "us-east-1b", "us-east-1c"])
  cidr_block = var.subnet_cidrs[each.key]
}

# Migrate state manually:
terraform state mv \
  'aws_subnet.private[0]' \
  'aws_subnet.private["us-east-1a"]'
terraform state mv \
  'aws_subnet.private[1]' \
  'aws_subnet.private["us-east-1b"]'
terraform state mv \
  'aws_subnet.private[2]' \
  'aws_subnet.private["us-east-1c"]'
```

**Import Existing Infrastructure:**
```hcl
# Terraform 1.5+ import block (declarative)
import {
  id = "i-0a1b2c3d4e5f678"
  to = aws_instance.web_server
}

# Import multiple resources
import {
  for_each = {
    "us-east-1a" = "subnet-111"
    "us-east-1b" = "subnet-222"
    "us-east-1c" = "subnet-333"
  }
  id = each.value
  to = aws_subnet.private[each.key]
}

# Generate config for existing resources (Terraform 1.5+)
# terraform plan -generate-config-out=generated.tf
```

```bash
# Traditional import (pre-1.5)
terraform import aws_instance.web_server i-0a1b2c3d4e5f678
terraform import aws_s3_bucket.data my-existing-bucket
terraform import aws_route53_record.www Z1234567890/example.com/A
terraform import 'aws_subnet.private["us-east-1a"]' subnet-12345

# Import entire module
terraform import module.vpc.aws_vpc.main vpc-0a1b2c3d
```

---

## ⚡ CATEGORY 7: Performance & Scalability

---

### 11. How Do You Scale Terraform for Large Infrastructure — Atlantis, Terragrunt, and Module Versioning?

**Expected Answer:**

**Terragrunt — DRY Terraform Configurations:**
```
infrastructure/
├── terragrunt.hcl              # Root config (backend, providers)
├── production/
│   ├── account.hcl             # Account-level config
│   ├── us-east-1/
│   │   ├── region.hcl          # Region-level config
│   │   ├── networking/
│   │   │   └── terragrunt.hcl
│   │   ├── eks/
│   │   │   └── terragrunt.hcl
│   │   └── rds/
│   │       └── terragrunt.hcl
│   └── eu-west-1/
└── staging/
```

```hcl
# Root terragrunt.hcl
locals {
  account_vars = read_terragrunt_config(find_in_parent_folders("account.hcl"))
  region_vars  = read_terragrunt_config(find_in_parent_folders("region.hcl"))

  account_id   = local.account_vars.locals.account_id
  aws_region   = local.region_vars.locals.aws_region
  environment  = local.account_vars.locals.environment
}

# Auto-configure backend for every module
remote_state {
  backend = "s3"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    bucket         = "company-terraform-state-${local.account_id}"
    key            = "${path_relative_to_include()}/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}

# Auto-generate provider for every module
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
provider "aws" {
  region = "${local.aws_region}"
  assume_role {
    role_arn = "arn:aws:iam::${local.account_id}:role/TerraformRole"
  }
  default_tags {
    tags = {
      Environment = "${local.environment}"
      ManagedBy   = "terragrunt"
    }
  }
}
EOF
}
```

```hcl
# production/us-east-1/eks/terragrunt.hcl
terraform {
  source = "git::https://github.com/company/terraform-modules.git//eks?ref=v3.2.1"
}

include "root" {
  path   = find_in_parent_folders()
  expose = true
}

# Reference output from another module (dependency management)
dependency "networking" {
  config_path = "../networking"

  # Mock outputs for plan without running dependencies
  mock_outputs = {
    vpc_id             = "vpc-fake"
    private_subnet_ids = ["subnet-fake1", "subnet-fake2"]
  }
  mock_outputs_allowed_terraform_commands = ["validate", "plan"]
}

dependency "iam" {
  config_path = "../iam"
}

inputs = {
  cluster_name       = "production-eks"
  kubernetes_version = "1.28"
  vpc_id             = dependency.networking.outputs.vpc_id
  subnet_ids         = dependency.networking.outputs.private_subnet_ids
  node_iam_role_arn  = dependency.iam.outputs.eks_node_role_arn

  node_groups = {
    general = {
      instance_types = ["m5.2xlarge"]
      min_size       = 3
      max_size       = 50
      desired_size   = 5
    }
  }
}
```

```bash
# Terragrunt commands
terragrunt run-all plan    # Plan all modules in directory tree
terragrunt run-all apply   # Apply all modules respecting dependencies
terragrunt run-all destroy # Destroy all in reverse dependency order

# Run for specific module
cd production/us-east-1/eks
terragrunt plan
terragrunt apply
```

**Atlantis — PR-based Automation:**
```yaml
# atlantis.yaml
version: 3

projects:
  - name: production-us-east-1-networking
    dir: environments/production/us-east-1/networking
    workspace: default
    terraform_version: v1.6.0
    autoplan:
      when_modified:
        - "*.tf"
        - "../../../modules/networking/**/*.tf"
      enabled: true
    apply_requirements:
      - approved           # Require PR approval
      - mergeable          # No merge conflicts
      - undiverged         # Branch up to date with base

  - name: production-eks
    dir: environments/production/us-east-1/eks
    workflow: production-workflow

workflows:
  production-workflow:
    plan:
      steps:
        - init:
            extra_args: ["-upgrade"]
        - run: tfsec . --no-color
        - run: tflint --format compact
        - plan:
            extra_args: ["-no-color", "-input=false"]
        - run: |
            infracost breakdown \
              --path=$PLANFILE \
              --format=table
    apply:
      steps:
        - apply:
            extra_args: ["-no-color", "-input=false", "-auto-approve"]
```

---

### 12. Explain Terraform Performance Optimization for Large Deployments

**Expected Answer:**

```bash
# ─── Parallelism Control ────────────────────────────────────
# Default parallelism is 10 concurrent operations
terraform apply -parallelism=30     # Increase for faster applies
terraform apply -parallelism=1      # Sequential (for debugging)

# ─── Targeted Operations ────────────────────────────────────
# Apply only specific resources (use sparingly)
terraform plan -target=module.eks
terraform apply -target=aws_eks_cluster.main

# Apply multiple targets
terraform apply \
  -target=module.networking \
  -target=module.security

# ─── Skip Refresh for Speed ─────────────────────────────────
# Skip refreshing state (use with caution)
terraform plan -refresh=false

# Refresh only (no changes)
terraform apply -refresh-only

# ─── State Splitting ────────────────────────────────────────
# Split large state files by layer
# Anti-pattern: one state for everything
# Best practice: separate states per layer

# Layer 1: Global resources (IAM, DNS) — changes rarely
# Layer 2: Networking (VPC, subnets) — changes rarely
# Layer 3: Platform (EKS, RDS) — changes sometimes
# Layer 4: Applications — changes frequently
```

```hcl
# ─── Data Source Optimization ───────────────────────────────
# Cache data sources at module level, pass as variables

# ❌ Each module fetches same AMI
module "app1" {
  source = "./modules/ec2"
  # Module internally fetches AMI data
}

module "app2" {
  source = "./modules/ec2"
  # Module internally fetches AMI again
}

# ✅ Fetch once, pass to modules
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

module "app1" {
  source = "./modules/ec2"
  ami_id = data.aws_ami.amazon_linux.id
}

module "app2" {
  source = "./modules/ec2"
  ami_id = data.aws_ami.amazon_linux.id
}
```

---

## 🧪 CATEGORY 8: Testing

---

### 13. How Do You Test Terraform Code? Explain Unit, Integration, and End-to-End Testing

**Expected Answer:**

**Testing Pyramid:**
```
         /\
        /  \         ← E2E Tests (Terratest)
       /    \          Deploy real infra, verify behavior
      /──────\
     /        \      ← Integration Tests (terraform test)
    /          \       Full module with mocked providers
   /────────────\
  /              \   ← Unit Tests (terraform test + mocks)
 /                \    Variables, outputs, logic validation
/──────────────────\
         ↑
  Static Analysis
  (tfsec, tflint, checkov)
```

**Static Analysis:**
```bash
# 1. Format check
terraform fmt -check -recursive

# 2. Validate syntax and internal consistency
terraform validate

# 3. tflint — Best practices linting
tflint --init
tflint --recursive

# 4. tfsec — Security scanning
tfsec . --severity HIGH

# 5. Checkov — Policy as code
checkov -d . \
  --framework terraform \
  --skip-check CKV_AWS_20,CKV_AWS_28

# 6. Infracost — Cost estimation
infracost breakdown --path .

# 7. terraform-docs — Documentation
terraform-docs markdown table --output-file README.md .
```

**Terraform Native Testing (v1.6+):**
```hcl
# tests/networking.tftest.hcl

# Mocked provider — no real AWS calls
mock_provider "aws" {
  mock_resource "aws_vpc" {
    defaults = {
      id         = "vpc-mock123"
      arn        = "arn:aws:ec2:us-east-1:123:vpc/vpc-mock123"
      cidr_block = "10.0.0.0/16"
    }
  }

  mock_resource "aws_subnet" {
    defaults = {
      id  = "subnet-mock123"
      arn = "arn:aws:ec2:us-east-1:123:subnet/subnet-mock123"
    }
  }
}

variables {
  environment        = "test"
  project            = "myapp"
  vpc_cidr           = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b"]
}

# Test 1: Validate basic creation
run "creates_vpc" {
  command = plan

  assert {
    condition     = aws_vpc.main.cidr_block == "10.0.0.0/16"
    error_message = "VPC CIDR block is incorrect"
  }

  assert {
    condition     = aws_vpc.main.enable_dns_hostnames == true
    error_message = "DNS hostnames must be enabled"
  }
}

# Test 2: Validate subnet count
run "creates_correct_subnets" {
  command = plan

  assert {
    condition     = length(aws_subnet.public) == 2
    error_message = "Expected 2 public subnets, one per AZ"
  }

  assert {
    condition     = length(aws_subnet.private) == 2
    error_message = "Expected 2 private subnets, one per AZ"
  }
}

# Test 3: Validate environment-based naming
run "validates_naming_convention" {
  command = plan

  assert {
    condition = alltrue([
      for subnet in aws_subnet.public :
      can(regex("^test-myapp-public-", subnet.tags["Name"]))
    ])
    error_message = "Subnets must follow naming convention: env-project-type-az"
  }
}

# Test 4: Validate tags
run "validates_required_tags" {
  command = plan

  assert {
    condition = alltrue([
      contains(keys(aws_vpc.main.tags), "Environment"),
      contains(keys(aws_vpc.main.tags), "ManagedBy"),
      contains(keys(aws_vpc.main.tags), "Module"),
    ])
    error_message = "VPC must have required tags: Environment, ManagedBy, Module"
  }
}
```

**Terratest — Integration/E2E Testing:**
```go
package test

import (
    "testing"
    "time"

    "github.com/gruntwork-io/terratest/modules/aws"
    "github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNetworkingModule(t *testing.T) {
    t.Parallel()

    awsRegion := "us-east-1"

    terraformOptions := &terraform.Options{
        // Path to the Terraform code to test
        TerraformDir: "../../modules/networking",

        Vars: map[string]interface{}{
            "environment":        "test",
            "project":            "terratest",
            "vpc_cidr":           "10.99.0.0/16",
            "availability_zones": []string{"us-east-1a", "us-east-1b"},
            "enable_nat_gateway": false,   // Save costs in tests
        },

        // Retry on eventual consistency issues
        RetryableTerraformErrors: map[string]string{
            "RequestError: send request failed": "Transient AWS error",
        },
        MaxRetries:         3,
        TimeBetweenRetries: 5 * time.Second,
    }

    // Always destroy after test
    defer terraform.Destroy(t, terraformOptions)

    // Deploy infrastructure
    terraform.InitAndApply(t, terraformOptions)

    // Validate outputs
    vpcID := terraform.Output(t, terraformOptions, "vpc_id")
    require.NotEmpty(t, vpcID)

    publicSubnetIDs := terraform.OutputList(t, terraformOptions, "public_subnet_ids")
    assert.Equal(t, 2, len(publicSubnetIDs), "Expected 2 public subnets")

    // Validate actual AWS resources via API
    vpc := aws.GetVpcById(t, vpcID, awsRegion)
    assert.Equal(t, "10.99.0.0/16", aws.GetVpcAttribute(t, vpc, "cidr_block"))
    assert.True(t, aws.GetVpcAttribute(t, vpc, "enable_dns_hostnames") == "true")

    // Validate subnet CIDR ranges
    for _, subnetID := range publicSubnetIDs {
        subnet := aws.GetSubnetById(t, subnetID, awsRegion)
        assert.Contains(t, subnet.CidrBlock, "10.99.")
    }

    // Validate tags
    vpcTags := aws.GetTagsForVpc(t, vpcID, awsRegion)
    assert.Equal(t, "test", vpcTags["Environment"])
    assert.Equal(t, "terraform", vpcTags["ManagedBy"])
}
```

```bash
# Run terratest
cd test
go test -v -timeout 30m -run TestNetworkingModule

# Run with specific test
go test -v -timeout 30m -run TestNetworkingModule/creates_vpc
```

---

## 🚨 CATEGORY 9: Troubleshooting

---

### 14. How Do You Debug and Troubleshoot Terraform Issues in Production?

**Expected Answer:**

```bash
# ─── Enable Debug Logging ────────────────────────────────────
export TF_LOG=DEBUG       # TRACE, DEBUG, INFO, WARN, ERROR
export TF_LOG_PATH=/tmp/terraform-debug.log
terraform plan 2>&1

# Log specific subsystems (Terraform 1.1+)
export TF_LOG_CORE=INFO
export TF_LOG_PROVIDER=DEBUG

# Disable logging
unset TF_LOG

# ─── Common Issues & Solutions ───────────────────────────────

# 1. State Lock Stuck
terraform force-unlock <LOCK_ID>
# Or directly remove DynamoDB lock item:
aws dynamodb delete-item \
  --table-name terraform-locks \
  --key '{"LockID": {"S": "company-state-bucket/prod/terraform.tfstate"}}'

# 2. Resource Already Exists (out-of-band creation)
terraform import aws_s3_bucket.existing my-existing-bucket

# 3. Drift Detection — infrastructure changed outside Terraform
terraform plan -refresh-only
terraform apply -refresh-only    # Update state to match reality

# 4. Dependency Cycle
terraform graph | dot -Tpng > graph.png   # Visualize to find cycle
# Look for circular references in configuration

# 5. Provider Version Conflicts
terraform providers lock \
  -platform=linux_amd64 \
  -platform=darwin_amd64 \
  -platform=windows_amd64

# 6. State File Corruption
# Pull current state
terraform state pull > state-backup.json

# Verify with JSON validation
cat state-backup.json | python3 -m json.tool > /dev/null

# Restore from S3 version
aws s3api list-object-versions \
  --bucket company-terraform-state \
  --key prod/terraform.tfstate

aws s3api get-object \
  --bucket company-terraform-state \
  --key prod/terraform.tfstate \
  --version-id <VERSION_ID> \
  state-restored.json

# 7. Plan shows unexpected destroy
terraform plan -out=plan.bin
terraform show -json plan.bin | \
  jq '.resource_changes[] | select(.change.actions[] == "delete") | .address'

# Check why resource is being destroyed
terraform show -json plan.bin | \
  jq '.resource_changes[] | select(.address == "aws_instance.app") | .change'
```

**Sensitive Plan Analysis:**
```bash
# Save plan and inspect
terraform plan -out=tfplan -no-color 2>&1 | tee plan.txt

# Convert plan to JSON for analysis
terraform show -json tfplan > plan.json

# Count changes by action type
cat plan.json | jq '
  .resource_changes |
  group_by(.change.actions[]) |
  map({
    action: .[0].change.actions[0],
    count: length,
    resources: [.[].address]
  })
'

# Find resources being replaced
cat plan.json | jq '
  [.resource_changes[] |
  select(.change.actions == ["delete", "create"] or
         .change.actions == ["create", "delete"]) |
  {address, reason: .action_reason}]
'
```

---

### 15. Design a Complete Terraform Architecture for a Multi-Account AWS Landing Zone

**Expected Answer:**

```
┌─────────────────────────────────────────────────────────────────────┐
│              AWS Multi-Account Landing Zone with Terraform          │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    Management Account                        │  │
│  │  ┌──────────┐  ┌──────────────┐  ┌────────────────────────┐ │  │
│  │  │   AWS    │  │  Terraform   │  │   IAM Identity Center  │ │  │
│  │  │ Control  │  │    Cloud /   │  │    (SSO)               │ │  │
│  │  │  Tower   │  │  Atlantis    │  │                        │ │  │
│  │  └──────────┘  └──────────────┘  └────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌──────────────────┐  ┌───────────────────┐  ┌────────────────┐  │
│  │  Security        │  │   Shared Services  │  │   Log Archive  │  │
│  │  Account         │  │   Account          │  │   Account      │  │
│  │                  │  │                    │  │                │  │
│  │  SecurityHub     │  │  Transit Gateway   │  │  CloudTrail    │  │
│  │  GuardDuty       │  │  DNS (Route53)     │  │  Config Logs   │  │
│  │  AWS Config      │  │  Shared ECR        │  │  VPC Flow Logs │  │
│  └──────────────────┘  └───────────────────┘  └────────────────┘  │
│                                                                     │
│  ┌──────────────────┐  ┌───────────────────┐  ┌────────────────┐  │
│  │  Production      │  │   Staging         │  │  Development   │  │
│  │  Account         │  │   Account         │  │  Account       │  │
│  │  111111111111    │  │   222222222222     │  │  333333333333  │  │
│  └──────────────────┘  └───────────────────┘  └────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

**Landing Zone Terraform Structure:**
```hcl
# ─── Root Module: Management Account ────────────────────────
# management/main.tf

module "organizations" {
  source = "./modules/organizations"

  accounts = {
    security = {
      name  = "company-security"
      email = "security@company.com"
      ou    = "security"
    }
    shared_services = {
      name  = "company-shared-services"
      email = "platform@company.com"
      ou    = "infrastructure"
    }
    production = {
      name  = "company-production"
      email = "production@company.com"
      ou    = "workloads"
    }
    staging = {
      name  = "company-staging"
      email = "staging@company.com"
      ou    = "workloads"
    }
  }

  # Service Control Policies (SCPs)
  scps = {
    deny_root_usage = {
      description = "Deny root account usage"
      policy      = file("policies/deny-root.json")
      targets     = ["ALL"]
    }
    restrict_regions = {
      description = "Restrict to approved regions"
      policy      = file("policies/restrict-regions.json")
      targets     = ["workloads", "infrastructure"]
    }
    deny_leave_organization = {
      description = "Prevent accounts leaving org"
      policy      = file("policies/deny-leave-org.json")
      targets     = ["ALL"]
    }
  }
}

# ─── Cross-Account IAM Roles ─────────────────────────────────
module "cross_account_roles" {
  source = "./modules/iam/cross-account"

  for_each = module.organizations.accounts

  target_account_id = each.value.account_id
  management_account_id = data.aws_caller_identity.current.account_id

  roles = {
    TerraformRole = {
      description  = "Terraform deployment role"
      policies     = ["arn:aws:iam::aws:policy/AdministratorAccess"]
      trust_github = true
      github_org   = "company"
      github_repo  = "terraform-infrastructure"
    }
    ReadOnlyRole = {
      description = "Read-only access for developers"
      policies    = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]
    }
  }
}

# ─── Shared Services Account ─────────────────────────────────
module "transit_gateway" {
  source    = "./modules/networking/transit-gateway"
  providers = { aws = aws.shared_services }

  name = "company-tgw"

  # Share with all accounts via RAM
  share_with_accounts = [
    module.organizations.accounts["production"].account_id,
    module.organizations.accounts["staging"].account_id,
  ]
}

module "shared_dns" {
  source    = "./modules/dns"
  providers = { aws = aws.shared_services }

  domain = "internal.company.com"

  # Share private hosted zone with all VPCs
  vpc_associations = {
    production = {
      vpc_id     = module.production_networking.vpc_id
      vpc_region = "us-east-1"
    }
    staging = {
      vpc_id     = module.staging_networking.vpc_id
      vpc_region = "us-east-1"
    }
  }
}

# ─── Security Account ────────────────────────────────────────
module "security_hub" {
  source    = "./modules/security/securityhub"
  providers = { aws = aws.security }

  member_accounts = values(module.organizations.accounts)[*].account_id

  standards = [
    "arn:aws:securityhub:us-east-1::standards/cis-aws-foundations-benchmark/v/1.4.0",
    "arn:aws:securityhub:us-east-1::standards/aws-foundational-security-best-practices/v/1.0.0"
  ]
}

module "guardduty" {
  source    = "./modules/security/guardduty"
  providers = { aws = aws.security }

  member_accounts     = module.organizations.accounts
  enable_s3_protection = true
  enable_eks_protection = true
  enable_malware_protection = true

  threat_intel_set_url = "s3://company-threat-intel/custom-indicators.txt"
}

# ─── Production Account ──────────────────────────────────────
module "production_networking" {
  source    = "./modules/networking"
  providers = { aws = aws.production }

  environment        = "production"
  project            = "company"
  vpc_cidr           = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]

  # Connect to Transit Gateway
  transit_gateway_id = module.transit_gateway.tgw_id
}

module "production_eks" {
  source    = "./modules/compute/eks"
  providers = { aws = aws.production }

  cluster_name       = "production"
  kubernetes_version = "1.28"
  vpc_id             = module.production_networking.vpc_id
  subnet_ids         = module.production_networking.private_subnet_ids

  node_groups = {
    system = {
      instance_types = ["m5.large"]
      min_size       = 2
      max_size       = 10
      taints = [{
        key    = "CriticalAddonsOnly"
        value  = "true"
        effect = "NO_SCHEDULE"
      }]
    }
    application = {
      instance_types = ["m5.2xlarge"]
      min_size       = 3
      max_size       = 100
      capacity_type  = "SPOT"
    }
    gpu = {
      instance_types = ["g4dn.xlarge"]
      min_size       = 0
      max_size       = 10
      capacity_type  = "ON_DEMAND"
    }
  }

  # Add-ons
  addons = {
    coredns            = "v1.10.1-eksbuild.4"
    kube-proxy         = "v1.28.2-eksbuild.2"
    vpc-cni            = "v1.15.4-eksbuild.1"
    aws-ebs-csi-driver = "v1.25.0-eksbuild.1"
    aws-efs-csi-driver = "v1.7.2-eksbuild.1"
  }
}
```

**Production-Grade tfvars:**
```hcl
# environments/production/us-east-1/terraform.tfvars

# ─── General ──────────────────────────────────────────────────
environment = "production"
region      = "us-east-1"
project     = "company"

# ─── Networking ───────────────────────────────────────────────
vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
enable_nat_gateway = true
single_nat_gateway = false    # HA: one NAT GW per AZ

# ─── EKS ──────────────────────────────────────────────────────
kubernetes_version       = "1.28"
cluster_endpoint_private = true
cluster_endpoint_public  = true
cluster_public_cidrs     = ["203.0.113.0/24"]  # VPN/Office CIDRs only

# ─── RDS ──────────────────────────────────────────────────────
db_instance_class    = "db.r5.2xlarge"
db_multi_az          = true
db_storage_encrypted = true
db_deletion_protection = true
db_backup_retention  = 30
db_performance_insights = true

# ─── Tags ─────────────────────────────────────────────────────
tags = {
  Team        = "platform"
  CostCenter  = "engineering"
  Owner       = "platform-team@company.com"
  Compliance  = "pci-dss"
  DataClass   = "confidential"
}
```

---

> 💡 **Senior Engineer Interview Tips:**
> - Always discuss **state management strategy** — it's the heart of Terraform
> - Show knowledge of **GitOps patterns** — Atlantis, Terraform Cloud, GitHub Actions
> - Reference **real war stories** — "We had a state corruption incident where..."
> - Demonstrate understanding of **blast radius** and why environment isolation matters
> - Know the **Terraform ecosystem** — Terragrunt, Terratest, tfsec, Checkov, Infracost
> - Discuss **cost implications** — always mention cost estimation in planning workflow
> - Show knowledge of **provider-specific nuances** — AWS, GCP, or Azure internals
> - Mention **compliance requirements** — how Terraform enforces SOC2/PCI-DSS controls via OPA/Sentinel
