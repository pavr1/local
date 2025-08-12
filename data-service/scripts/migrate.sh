#!/bin/bash

# Ice Cream Store Database Migration Script
# This script handles database migrations for the application

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="$PROJECT_ROOT/docker/init/migrations"

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
RESET='\033[0m'

# Database connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="icecream_store"
DB_USER="postgres"
DB_PASSWORD="postgres123"

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  up [N]     Apply N migrations (default: all pending)"
    echo "  down [N]   Rollback N migrations (default: 1)"
    echo "  status     Show migration status"
    echo "  create     Create new migration files"
    echo ""
    echo "Examples:"
    echo "  $0 up              # Apply all pending migrations"
    echo "  $0 up 2            # Apply next 2 migrations"
    echo "  $0 down            # Rollback last migration"
    echo "  $0 down 3          # Rollback last 3 migrations"
    echo "  $0 status          # Show current migration status"
    echo ""
}

# Function to check if database is accessible
check_database() {
    echo -e "${CYAN}🔍 Checking database connection...${RESET}"
    
    if ! docker exec icecream_postgres pg_isready -U postgres -d icecream_store > /dev/null 2>&1; then
        echo -e "${RED}❌ Database is not accessible. Make sure the database is running.${RESET}"
        echo "   Run: make start (in data-service directory)"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Database connection verified${RESET}"
}

# Function to create migrations table if it doesn't exist
create_migrations_table() {
    echo -e "${CYAN}📋 Setting up migrations table...${RESET}"
    
    docker exec icecream_postgres psql -U postgres -d icecream_store -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    " > /dev/null 2>&1
    
    echo -e "${GREEN}✅ Migrations table ready${RESET}"
}

# Function to get applied migrations
get_applied_migrations() {
    docker exec icecream_postgres psql -U postgres -d icecream_store -t -c "
        SELECT version FROM schema_migrations ORDER BY version;
    " 2>/dev/null | grep -v '^$' | tr -d ' ' || echo ""
}

# Function to get pending migrations
get_pending_migrations() {
    local applied_migrations="$1"
    local pending_migrations=""
    
    for migration_file in "$MIGRATIONS_DIR"/*.up.sql; do
        if [[ -f "$migration_file" ]]; then
            local version=$(basename "$migration_file" .up.sql)
            if ! echo "$applied_migrations" | grep -q "^$version$"; then
                pending_migrations="$pending_migrations $version"
            fi
        fi
    done
    
    echo "$pending_migrations" | tr ' ' '\n' | grep -v '^$' | sort
}

# Function to apply migration
apply_migration() {
    local version="$1"
    local up_file="$MIGRATIONS_DIR/${version}.up.sql"
    
    if [[ ! -f "$up_file" ]]; then
        echo -e "${RED}❌ Migration file not found: $up_file${RESET}"
        return 1
    fi
    
    echo -e "${CYAN}📈 Applying migration: $version${RESET}"
    
    # Copy migration file to container
    docker cp "$up_file" icecream_postgres:/tmp/migration.sql
    
    # Apply the migration
    if docker exec icecream_postgres psql -U postgres -d icecream_store -f /tmp/migration.sql > /dev/null 2>&1; then
        # Record the migration
        docker exec icecream_postgres psql -U postgres -d icecream_store -c "
            INSERT INTO schema_migrations (version) VALUES ('$version');
        " > /dev/null 2>&1
        
        echo -e "${GREEN}✅ Migration applied: $version${RESET}"
        return 0
    else
        echo -e "${RED}❌ Failed to apply migration: $version${RESET}"
        return 1
    fi
}

# Function to rollback migration
rollback_migration() {
    local version="$1"
    local down_file="$MIGRATIONS_DIR/${version}.down.sql"
    
    if [[ ! -f "$down_file" ]]; then
        echo -e "${RED}❌ Rollback file not found: $down_file${RESET}"
        return 1
    fi
    
    echo -e "${CYAN}📉 Rolling back migration: $version${RESET}"
    
    # Copy rollback file to container
    docker cp "$down_file" icecream_postgres:/tmp/rollback.sql
    
    # Apply the rollback
    if docker exec icecream_postgres psql -U postgres -d icecream_store -f /tmp/rollback.sql > /dev/null 2>&1; then
        # Remove the migration record
        docker exec icecream_postgres psql -U postgres -d icecream_store -c "
            DELETE FROM schema_migrations WHERE version = '$version';
        " > /dev/null 2>&1
        
        echo -e "${GREEN}✅ Migration rolled back: $version${RESET}"
        return 0
    else
        echo -e "${RED}❌ Failed to rollback migration: $version${RESET}"
        return 1
    fi
}

# Function to show migration status
show_status() {
    echo -e "${CYAN}📊 Migration Status${RESET}"
    echo "=================="
    
    local applied_migrations=$(get_applied_migrations)
    local pending_migrations=$(get_pending_migrations "$applied_migrations")
    
    echo -e "${GREEN}✅ Applied Migrations:${RESET}"
    if [[ -z "$applied_migrations" ]]; then
        echo "   None"
    else
        echo "$applied_migrations" | while read -r version; do
            echo "   $version"
        done
    fi
    
    echo ""
    echo -e "${YELLOW}⏳ Pending Migrations:${RESET}"
    if [[ -z "$pending_migrations" ]]; then
        echo "   None (database is up to date)"
    else
        echo "$pending_migrations" | while read -r version; do
            echo "   $version"
        done
    fi
    
    echo ""
}

# Function to create new migration
create_migration() {
    local description="$1"
    
    if [[ -z "$description" ]]; then
        echo -e "${RED}❌ Please provide a description for the migration${RESET}"
        echo "Usage: $0 create <description>"
        exit 1
    fi
    
    # Get next migration number
    local next_number=1
    for migration_file in "$MIGRATIONS_DIR"/*.up.sql; do
        if [[ -f "$migration_file" ]]; then
            local current_number=$(basename "$migration_file" | cut -d'_' -f1)
            if [[ "$current_number" -ge "$next_number" ]]; then
                next_number=$((current_number + 1))
            fi
        fi
    done
    
    # Format the number with leading zeros
    local formatted_number=$(printf "%03d" $next_number)
    local filename="${formatted_number}_${description// /_}"
    
    # Create up migration
    cat > "$MIGRATIONS_DIR/${filename}.up.sql" << EOF
-- Migration: $description
-- UP Migration: Add your migration SQL here

-- Example:
-- ALTER TABLE table_name ADD COLUMN column_name TYPE;

EOF
    
    # Create down migration
    cat > "$MIGRATIONS_DIR/${filename}.down.sql" << EOF
-- Migration: $description
-- DOWN Migration: Rollback your migration SQL here

-- Example:
-- ALTER TABLE table_name DROP COLUMN column_name;

EOF
    
    echo -e "${GREEN}✅ Created migration files:${RESET}"
    echo "   $MIGRATIONS_DIR/${filename}.up.sql"
    echo "   $MIGRATIONS_DIR/${filename}.down.sql"
}

# Main script logic
main() {
    local command="$1"
    local count="$2"
    
    case "$command" in
        "up")
            check_database
            create_migrations_table
            
            local applied_migrations=$(get_applied_migrations)
            local pending_migrations=$(get_pending_migrations "$applied_migrations")
            
            if [[ -z "$pending_migrations" ]]; then
                echo -e "${GREEN}✅ No pending migrations${RESET}"
                return 0
            fi
            
            local migrations_to_apply="$pending_migrations"
            if [[ -n "$count" && "$count" -gt 0 ]]; then
                migrations_to_apply=$(echo "$pending_migrations" | head -n "$count")
            fi
            
            echo -e "${CYAN}🚀 Applying migrations...${RESET}"
            echo "$migrations_to_apply" | while read -r version; do
                if ! apply_migration "$version"; then
                    echo -e "${RED}❌ Migration failed. Stopping.${RESET}"
                    exit 1
                fi
            done
            
            echo -e "${GREEN}🎉 All migrations completed successfully!${RESET}"
            ;;
            
        "down")
            check_database
            create_migrations_table
            
            local applied_migrations=$(get_applied_migrations)
            if [[ -z "$applied_migrations" ]]; then
                echo -e "${GREEN}✅ No applied migrations to rollback${RESET}"
                return 0
            fi
            
            local migrations_to_rollback=$(echo "$applied_migrations" | tail -n "${count:-1}")
            
            echo -e "${CYAN}🔄 Rolling back migrations...${RESET}"
            echo "$migrations_to_rollback" | tail -r | while read -r version; do
                if ! rollback_migration "$version"; then
                    echo -e "${RED}❌ Rollback failed. Stopping.${RESET}"
                    exit 1
                fi
            done
            
            echo -e "${GREEN}🎉 Rollback completed successfully!${RESET}"
            ;;
            
        "status")
            check_database
            create_migrations_table
            show_status
            ;;
            
        "create")
            create_migration "$count"
            ;;
            
        *)
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 