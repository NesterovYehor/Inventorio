package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/NesterovYehor/Inventorio/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *DB {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Connect(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			assert.NoError(t, err)
		}
	})
	return db
}

func TestConnectCreatesSchema(t *testing.T) {
	db := testDB(t)
	require.NotNil(t, db)
	require.NotNil(t, db.conn)
}

func TestAddProperty(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	id, err := db.AddProperty(ctx)
	require.NoError(t, err)
	require.Greater(t, id, int64(0), "expected positive property id")

	needs, err := db.GetNeedsByPropertyID(ctx, id)
	require.NoError(t, err)
	require.Empty(t, needs, "expected no needs for empty items table")
}

func TestAddPropertySeedsNeedsForExistingItems(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	_, err := db.AddNewItem(ctx)
	require.NoError(t, err)
	_, err = db.AddNewItem(ctx)
	require.NoError(t, err)

	id, err := db.AddProperty(ctx)
	require.NoError(t, err)

	needs, err := db.GetNeedsByPropertyID(ctx, id)
	require.NoError(t, err)
	require.Len(t, needs, 2, "expected 2 needs (one per existing item)")
	for _, n := range needs {
		assert.Equal(t, int(id), n.PropertyID, "need property id mismatch")
		assert.Equal(t, 0, n.Quantity, "expected default quantity 0")
	}
}

func TestGetNeedsByPropertyID(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	propID, err := db.AddProperty(ctx)
	require.NoError(t, err)

	needs, err := db.GetNeedsByPropertyID(ctx, propID)
	require.NoError(t, err)
	require.Empty(t, needs, "expected no needs initially")

	_, err = db.AddNewItem(ctx)
	require.NoError(t, err)

	needs, err = db.GetNeedsByPropertyID(ctx, propID)
	require.NoError(t, err)
	require.Len(t, needs, 1, "expected 1 need after adding an item")
	assert.Equal(t, 1, needs[0].ItemID)
}

func TestAddNewItem(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	item, err := db.AddNewItem(ctx)
	require.NoError(t, err)
	require.Greater(t, item.ID, 0, "expected positive item id")
	assert.Empty(t, item.Name, "expected empty default name")
	assert.Equal(t, 0, item.Quantity, "expected default quantity 0")
}

func TestAddNewItemSeedsNeedsForExistingItems(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	// Add a property first; then adding items should create needs for it.
	propID, err := db.AddProperty(ctx)
	require.NoError(t, err)

	_, err = db.AddNewItem(ctx)
	require.NoError(t, err)

	needs, err := db.GetNeedsByPropertyID(ctx, propID)
	require.NoError(t, err)
	require.Len(t, needs, 1, "expected 1 need after adding item")
	assert.Equal(t, 1, needs[0].ItemID)

	_, err = db.AddNewItem(ctx)
	require.NoError(t, err)

	needs, err = db.GetNeedsByPropertyID(ctx, propID)
	require.NoError(t, err)
	require.Len(t, needs, 2, "expected 2 needs after adding second item")
}

func TestUpdateItemField(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	item, err := db.AddNewItem(ctx)
	require.NoError(t, err)

	require.NoError(t, db.UpdateItemField(ctx, "name", "Big Towels", int64(item.ID)))
	require.NoError(t, db.UpdateItemField(ctx, "quantity", 42, int64(item.ID)))

	items, err := db.GetAllItems(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1, "expected 1 item")
	assert.Equal(t, "Big Towels", items[0].Name)
	assert.Equal(t, 42, items[0].Quantity)
}

func TestUpdateItemFieldInvalidColumn(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	require.Error(t, db.UpdateItemField(ctx, "hacker", "x", 1), "expected error for invalid column")
}

func TestGetAllItems(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	items, err := db.GetAllItems(ctx)
	require.NoError(t, err)
	require.Empty(t, items, "expected empty items initially")

	_, err = db.AddNewItem(ctx)
	require.NoError(t, err)
	_, err = db.AddNewItem(ctx)
	require.NoError(t, err)

	items, err = db.GetAllItems(ctx)
	require.NoError(t, err)
	require.Len(t, items, 2, "expected 2 items")
	for _, it := range items {
		assert.Greater(t, it.ID, 0, "expected positive id")
		assert.Equal(t, 0, it.Quantity, "expected default quantity 0")
	}
}

func TestPropertyNeedModel(t *testing.T) {
	_ = models.PropertyNeed{}
}

func TestUpdateOrderExtraValue(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()

	id, err := db.AddNewOrder(ctx)

	require.NoError(t, err)
	require.NotEqual(t, 0, id)

	_, err = db.AddProperty(ctx)
	require.NoError(t, err)

	item, err := db.AddNewItem(ctx)
	require.NoError(t, err)

	require.NoError(t, db.UpdateItemField(ctx, "quantity", 5, int64(item.ID)))

	require.NoError(t, db.AddPropertyToOrder(ctx, "Unnamed Property", nil))

	require.NoError(t, db.UpdateOrderExtraValue(ctx, id, item.ID, 10))

	row, err := db.GetOrderItemRequirement(ctx, id, item.ID)
	require.NoError(t, err)
	require.Equal(t, 10, row.Extra)
	require.Equal(t, 5, row.OrderQty)

}
