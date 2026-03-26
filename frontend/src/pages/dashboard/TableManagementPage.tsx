import { useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useTournament } from '@/api/hooks/useTournaments';
import {
  useListTables,
  useCreateTables,
  useUpdateTable,
  useDeleteTable,
  useUploadPhoto,
  useDeletePhoto,
  useGenerateQRCode,
  useGenerateAllQRCodes,
  type TableSummary,
} from '@/api/hooks/useTables';

export function TableManagementPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data: tournament } = useTournament(slug);
  const { data: tables, isLoading } = useListTables(slug);
  const createTables = useCreateTables(slug);
  const exportPDF = useGenerateAllQRCodes(slug);

  if (isLoading) {
    return <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>{t('common.loading')}</p>;
  }

  const tableCount = tables?.length ?? 0;
  const targetCount = tournament?.table_count ?? 0;
  const isDraft = tournament?.status === 'draft';
  const missingCount = Math.max(0, targetCount - tableCount);
  const hasTables = tableCount > 0;

  return (
    <div className="space-y-8 max-w-3xl">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <Link
            to={`/dashboard/tournaments/${slug}`}
            className="text-xs hover:underline mb-1 block"
            style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
          >
            ← {t('table_mgmt.back_to_tournament')}
          </Link>
          <h1 className="font-heading text-2xl font-bold">
            {t('table_mgmt.title', { name: tournament?.name ?? slug })}
          </h1>
        </div>

        {hasTables && (
          <button
            onClick={() => exportPDF.mutate()}
            disabled={exportPDF.isPending}
            className="text-sm px-4 py-2 rounded border font-medium"
            style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
          >
            {exportPDF.isPending ? t('common.loading') : t('table_mgmt.qr_export_all')}
          </button>
        )}
      </div>

      {/* Create tables CTA — no tables yet */}
      {!hasTables && isDraft && (
        <div
          className="p-6 rounded text-center space-y-3"
          style={{ backgroundColor: 'var(--color-surface)' }}
        >
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('table_mgmt.create_tables_hint', { count: targetCount })}
          </p>
          <button
            onClick={() => createTables.mutate()}
            disabled={createTables.isPending}
            className="px-6 py-2 rounded text-sm font-medium"
            style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
          >
            {createTables.isPending ? t('common.loading') : t('table_mgmt.create_tables')}
          </button>
        </div>
      )}

      {/* Add missing tables CTA — some tables exist but count < target */}
      {hasTables && missingCount > 0 && isDraft && (
        <div className="flex items-center gap-3 p-3 rounded" style={{ backgroundColor: 'var(--color-surface)' }}>
          <p className="text-sm flex-1" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('table_mgmt.add_missing_tables', { count: missingCount })}
          </p>
          <button
            onClick={() => createTables.mutate()}
            disabled={createTables.isPending}
            className="shrink-0 px-4 py-1.5 rounded text-sm font-medium"
            style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
          >
            {createTables.isPending ? t('common.loading') : t('table_mgmt.create_tables')}
          </button>
        </div>
      )}

      {/* Table list */}
      {hasTables && (
        <div className="space-y-4">
          {tables.map((table) => (
            <TableCard key={table.id} slug={slug} table={table} isDraft={isDraft} />
          ))}
        </div>
      )}

    </div>
  );
}

// ── TableCard ──────────────────────────────────────────────────────────────

function TableCard({ slug, table, isDraft }: { slug: string; table: TableSummary; isDraft: boolean }) {
  const { t } = useTranslation();
  const updateTable = useUpdateTable(slug, table.number);
  const deleteTable = useDeleteTable(slug);
  const uploadPhoto = useUploadPhoto(slug, table.number);
  const deletePhoto = useDeletePhoto(slug);
  const generateQR = useGenerateQRCode(slug, table.number);

  const [name, setName] = useState(table.name ?? '');
  const [description, setDescription] = useState(table.description ?? '');
  const [dirty, setDirty] = useState(false);
  const fileRefs = {
    general: useRef<HTMLInputElement>(null),
    zone_a:  useRef<HTMLInputElement>(null),
    zone_b:  useRef<HTMLInputElement>(null),
  };

  const handleSave = () => {
    updateTable.mutate({ name: name || null, description: description || null });
    setDirty(false);
  };

  const handleUpload = async (category: 'general' | 'zone_a' | 'zone_b', file: File) => {
    const fd = new FormData();
    fd.append('photo', file);
    fd.append('category', category);
    await uploadPhoto.mutateAsync(fd);
  };

  const categories: Array<{ key: 'general' | 'zone_a' | 'zone_b'; label: string }> = [
    { key: 'general', label: t('photo.category_general') },
    { key: 'zone_a',  label: t('photo.category_zone_a') },
    { key: 'zone_b',  label: t('photo.category_zone_b') },
  ];

  return (
    <div
      className="rounded border p-4 space-y-4"
      style={{
        backgroundColor: 'var(--color-surface)',
        borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)',
      }}
    >
      {/* Title row */}
      <div className="flex items-center justify-between gap-2 flex-wrap">
        <h2 className="font-heading font-bold text-lg">
          {t('table_mgmt.table_number', { number: table.number })}
          {table.name && <span className="ml-2 text-base font-normal">— {table.name}</span>}
        </h2>

        <div className="flex gap-2">
          {isDraft && (
            <button
              onClick={() => {
                if (window.confirm(t('table_mgmt.delete_table_confirm'))) {
                  deleteTable.mutate(table.number);
                }
              }}
              disabled={deleteTable.isPending}
              className="text-xs px-3 py-1 rounded border font-medium"
              style={{ borderColor: '#dc2626', color: '#dc2626' }}
              aria-label={`${t('table_mgmt.delete_table')} ${table.number}`}
            >
              {t('table_mgmt.delete_table')}
            </button>
          )}
          {table.qr_code_url ? (
            <a
              href={table.qr_code_url}
              target="_blank"
              rel="noreferrer"
              className="text-xs px-3 py-1 rounded border font-medium"
              style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
            >
              {t('table_mgmt.qr_download')}
            </a>
          ) : (
            <button
              onClick={() => generateQR.mutate()}
              disabled={generateQR.isPending}
              className="text-xs px-3 py-1 rounded border font-medium"
              style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
            >
              {generateQR.isPending ? t('table_mgmt.qr_generating') : t('table_mgmt.qr_generate')}
            </button>
          )}
        </div>
      </div>

      {/* Name + description fields */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div>
          <label className="block text-xs font-medium mb-1">{t('table_mgmt.name_label')}</label>
          <input
            type="text"
            value={name}
            onChange={(e) => { setName(e.target.value); setDirty(true); }}
            placeholder={t('table_mgmt.name_placeholder')}
            className="w-full px-3 py-1.5 rounded text-sm border"
            style={{
              backgroundColor: 'var(--color-background)',
              borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
              color: 'var(--color-text)',
            }}
          />
        </div>
        <div>
          <label className="block text-xs font-medium mb-1">{t('table_mgmt.description_label')}</label>
          <input
            type="text"
            value={description}
            onChange={(e) => { setDescription(e.target.value); setDirty(true); }}
            placeholder=""
            className="w-full px-3 py-1.5 rounded text-sm border"
            style={{
              backgroundColor: 'var(--color-background)',
              borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
              color: 'var(--color-text)',
            }}
          />
        </div>
      </div>

      {dirty && (
        <button
          onClick={handleSave}
          disabled={updateTable.isPending}
          className="text-xs px-4 py-1.5 rounded font-medium"
          style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
        >
          {updateTable.isPending ? t('common.loading') : t('table_mgmt.save')}
        </button>
      )}

      {/* Photos by category */}
      <div className="space-y-3">
        {categories.map(({ key, label }) => {
          const photos = table.photos.filter((p) => p.category === key);
          return (
            <div key={key}>
              <p className="text-xs font-medium mb-1.5">{label}</p>
              <div className="flex flex-wrap gap-2 items-center">
                {photos.map((photo) => (
                  <div key={photo.id} className="relative group">
                    <img
                      src={photo.thumbnail_url ?? photo.url}
                      alt=""
                      className="h-16 w-16 object-cover rounded"
                    />
                    <button
                      onClick={() => {
                        if (window.confirm(t('photo.delete_confirm'))) {
                          deletePhoto.mutate(photo.id.toString());
                        }
                      }}
                      className="absolute -top-1.5 -right-1.5 h-5 w-5 rounded-full text-xs flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                      style={{ backgroundColor: '#dc2626', color: 'white' }}
                      aria-label="Delete photo"
                    >
                      ×
                    </button>
                  </div>
                ))}

                {/* Upload button */}
                <button
                  onClick={() => fileRefs[key].current?.click()}
                  disabled={uploadPhoto.isPending}
                  className="h-16 w-16 rounded border-2 border-dashed text-xs flex items-center justify-center"
                  style={{ borderColor: 'color-mix(in srgb, var(--color-primary) 40%, transparent)', color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}
                  aria-label={t('photo.tap_to_upload')}
                >
                  {uploadPhoto.isPending ? '…' : '+'}
                </button>
                <input
                  ref={fileRefs[key]}
                  type="file"
                  accept="image/*"
                  className="sr-only"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) handleUpload(key, file);
                    e.target.value = '';
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

