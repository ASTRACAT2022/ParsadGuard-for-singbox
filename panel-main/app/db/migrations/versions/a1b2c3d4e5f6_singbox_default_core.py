"""switch core_configs default to singbox + rename xray_version to core_version

Revision ID: a1b2c3d4e5f6
Revises: f02194c811d6
Create Date: 2026-05-16 00:00:00.000000

"""

from alembic import op


revision = "a1b2c3d4e5f6"
down_revision = "f02194c811d6"
branch_labels = None
depends_on = None


def upgrade() -> None:
    bind = op.get_bind()
    dialect = bind.dialect.name
    # Rename column on nodes table
    if dialect == "postgresql":
        op.execute("ALTER TABLE nodes RENAME COLUMN xray_version TO core_version")
        op.execute("ALTER TABLE core_configs ALTER COLUMN type SET DEFAULT 'singbox'")
    elif dialect in ("mysql", "mariadb"):
        op.execute("ALTER TABLE nodes RENAME COLUMN xray_version TO core_version")
        op.execute("ALTER TABLE core_configs ALTER COLUMN type SET DEFAULT 'singbox'")
    else:
        # SQLite: need to use batch mode for rename
        with op.batch_alter_table("nodes") as batch_op:
            batch_op.alter_column("xray_version", new_column_name="core_version")
            # SQLite doesn't support altering defaults; handled by SQLA on insert


def downgrade() -> None:
    bind = op.get_bind()
    dialect = bind.dialect.name
    if dialect == "postgresql":
        op.execute("ALTER TABLE nodes RENAME COLUMN core_version TO xray_version")
        op.execute("ALTER TABLE core_configs ALTER COLUMN type SET DEFAULT 'xray'")
    elif dialect in ("mysql", "mariadb"):
        op.execute("ALTER TABLE nodes RENAME COLUMN core_version TO xray_version")
        op.execute("ALTER TABLE core_configs ALTER COLUMN type SET DEFAULT 'xray'")
    else:
        with op.batch_alter_table("nodes") as batch_op:
            batch_op.alter_column("core_version", new_column_name="xray_version")
