import React, { useEffect, useState } from 'react';
import * as categoriesApi from '../api/categories';
import type { Category, CategoryType } from '../api/categories';

const AVAILABLE_ICONS = ['tag', 'shopping-cart', 'utensils', 'car', 'home', 'film', 'zap', 'briefcase', 'wallet', 'gift', 'heart', 'book'];
const AVAILABLE_COLORS = ['#6366F1', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#14B8A6'];

export const CategoriesManager: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  // Tab filter: 'ALL' | 'EXPENSE' | 'INCOME'
  const [activeTab, setActiveTab] = useState<'ALL' | 'EXPENSE' | 'INCOME'>('ALL');

  // Form states
  const [isFormOpen, setIsFormOpen] = useState<boolean>(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [name, setName] = useState<string>('');
  const [type, setType] = useState<CategoryType>('EXPENSE');
  const [icon, setIcon] = useState<string>('tag');
  const [color, setColor] = useState<string>('#6366F1');
  const [formError, setFormError] = useState<string>('');

  const loadCategories = async () => {
    setIsLoading(true);
    setError('');
    try {
      const data = await categoriesApi.listCategories();
      setCategories(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load categories');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadCategories();
  }, []);

  const resetForm = () => {
    setEditingId(null);
    setName('');
    setType('EXPENSE');
    setIcon('tag');
    setColor('#6366F1');
    setFormError('');
    setIsFormOpen(false);
  };

  const handleOpenCreate = () => {
    resetForm();
    setIsFormOpen(true);
  };

  const handleOpenEdit = (cat: Category) => {
    setEditingId(cat.id);
    setName(cat.name);
    setType(cat.type);
    setIcon(cat.icon || 'tag');
    setColor(cat.color || '#6366F1');
    setFormError('');
    setIsFormOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    const cleanName = name.trim();
    if (!cleanName) {
      setFormError('Category name is required');
      return;
    }

    try {
      if (editingId) {
        await categoriesApi.updateCategory(editingId, {
          name: cleanName,
          type,
          icon,
          color,
        });
        setSuccess('Category updated successfully!');
      } else {
        await categoriesApi.createCategory({
          name: cleanName,
          type,
          icon,
          color,
        });
        setSuccess('Category created successfully!');
      }
      resetForm();
      await loadCategories();
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Operation failed');
    }
  };

  const handleDelete = async (id: string, catName: string) => {
    if (!window.confirm(`Are you sure you want to delete category "${catName}"?`)) {
      return;
    }
    try {
      await categoriesApi.deleteCategory(id);
      setSuccess(`Category "${catName}" deleted.`);
      await loadCategories();
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete category');
    }
  };

  const filteredCategories = categories.filter((c) => {
    if (activeTab === 'ALL') return true;
    return c.type === activeTab;
  });

  return (
    <div className="categories-manager" id="categories-manager">
      <div className="section-header">
        <div>
          <h3 className="section-title">🏷️ Categories</h3>
          <p className="section-subtitle">
            Organize and classify your income streams and expense destinations.
          </p>
        </div>
        <button
          className="btn btn--primary"
          id="btn-add-category"
          onClick={handleOpenCreate}
        >
          + Add Category
        </button>
      </div>

      {success && (
        <div className="auth-alert auth-alert--success" role="status">
          <span>✅ {success}</span>
        </div>
      )}

      {error && (
        <div className="auth-alert auth-alert--error" role="alert">
          <span>⚠️ {error}</span>
        </div>
      )}

      {/* Filter Tabs */}
      <div className="tabs-container">
        <button
          className={`tab-btn ${activeTab === 'ALL' ? 'tab-btn--active' : ''}`}
          onClick={() => setActiveTab('ALL')}
        >
          All ({categories.length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'EXPENSE' ? 'tab-btn--active' : ''}`}
          onClick={() => setActiveTab('EXPENSE')}
        >
          💸 Expenses ({categories.filter((c) => c.type === 'EXPENSE').length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'INCOME' ? 'tab-btn--active' : ''}`}
          onClick={() => setActiveTab('INCOME')}
        >
          💰 Income ({categories.filter((c) => c.type === 'INCOME').length})
        </button>
      </div>

      {/* Create / Edit Form */}
      {isFormOpen && (
        <div className="form-card" id="category-form-card">
          <h4 className="form-card__title">
            {editingId ? 'Edit Category' : 'Create New Category'}
          </h4>

          {formError && (
            <div className="auth-alert auth-alert--error">
              <span>⚠️ {formError}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate>
            <div className="form-row">
              <div className="form-group flex-1">
                <label htmlFor="cat-name">Category Name</label>
                <input
                  id="cat-name"
                  type="text"
                  placeholder="e.g. Food & Dining, Salary"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                />
              </div>

              <div className="form-group flex-1">
                <label htmlFor="cat-type">Type</label>
                <select
                  id="cat-type"
                  className="form-select"
                  value={type}
                  onChange={(e) => setType(e.target.value as CategoryType)}
                >
                  <option value="EXPENSE">💸 Expense</option>
                  <option value="INCOME">💰 Income</option>
                </select>
              </div>
            </div>

            <div className="form-row">
              <div className="form-group flex-1">
                <label>Color Accent</label>
                <div className="color-palette">
                  {AVAILABLE_COLORS.map((c) => (
                    <button
                      key={c}
                      type="button"
                      className={`color-dot ${color === c ? 'color-dot--selected' : ''}`}
                      style={{ backgroundColor: c }}
                      onClick={() => setColor(c)}
                      title={c}
                    />
                  ))}
                </div>
              </div>

              <div className="form-group flex-1">
                <label htmlFor="cat-icon">Icon Identifier</label>
                <select
                  id="cat-icon"
                  className="form-select"
                  value={icon}
                  onChange={(e) => setIcon(e.target.value)}
                >
                  {AVAILABLE_ICONS.map((ic) => (
                    <option key={ic} value={ic}>{ic}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="form-actions">
              <button
                type="button"
                className="btn btn--outline"
                id="btn-cancel-category"
                onClick={resetForm}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn btn--primary"
                id="btn-submit-category"
              >
                {editingId ? 'Save Changes' : 'Create Category'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Categories List */}
      {isLoading ? (
        <div className="app-loading">
          <div className="spinner" />
          <span>Loading categories…</span>
        </div>
      ) : filteredCategories.length === 0 ? (
        <div className="empty-state" id="categories-empty-state">
          <span className="empty-icon">🏷️</span>
          <p className="empty-title">No Categories Found</p>
          <p className="empty-subtitle">
            Click "+ Add Category" to classify your finances.
          </p>
        </div>
      ) : (
        <div className="categories-grid" id="categories-list">
          {filteredCategories.map((cat) => (
            <div
              key={cat.id}
              className="category-card"
              id={`category-card-${cat.id}`}
              style={{ borderLeftColor: cat.color }}
            >
              <div className="category-meta">
                <span
                  className="category-badge-dot"
                  style={{ backgroundColor: cat.color }}
                />
                <div>
                  <h4 className="category-name">{cat.name}</h4>
                  <span className="category-type-tag">{cat.type}</span>
                </div>
              </div>

              <div className="category-actions">
                <button
                  className="btn-action"
                  onClick={() => handleOpenEdit(cat)}
                  id={`btn-edit-category-${cat.id}`}
                  title="Edit Category"
                >
                  ✏️
                </button>
                <button
                  className="btn-action btn-action--danger"
                  onClick={() => handleDelete(cat.id, cat.name)}
                  id={`btn-delete-category-${cat.id}`}
                  title="Delete Category"
                >
                  🗑️
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
