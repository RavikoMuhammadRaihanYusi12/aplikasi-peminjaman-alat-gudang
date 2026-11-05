// Preview Application Store with LocalStorage Persistence
const DEFAULT_TECHNICIANS = [
  { id: 1, name: 'Budi', section: 'Listrik', active: 1 },
  { id: 2, name: 'Andi', section: 'Mekanik', active: 1 },
  { id: 3, name: 'Candra', section: 'Listrik', active: 1 },
  { id: 4, name: 'Dedi', section: 'Mekanik', active: 1 },
  { id: 5, name: 'Eko', section: 'Utilitas', active: 1 },
  { id: 6, name: 'Fajar', section: 'Produksi', active: 1 },
  { id: 7, name: 'Gunawan', section: 'Listrik', active: 1 },
  { id: 8, name: 'Hendra', section: 'Mekanik', active: 1 },
  { id: 9, name: 'Irwan', section: 'Utilitas', active: 1 },
  { id: 10, name: 'Joko', section: 'Mekanik', active: 1 },
  { id: 11, name: 'Krisna', section: 'Listrik', active: 1 },
  { id: 12, name: 'Lukman', section: 'Produksi', active: 1 }
];

const DEFAULT_TOOLS = [
  { id: 1, code: 'MM-01', name: 'Multimeter Digital Fluke', category: 'Alat ukur', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 2, code: 'MM-02', name: 'Multimeter Sanwa', category: 'Alat ukur', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 3, code: 'MM-03', name: 'Multimeter Kyoritsu', category: 'Alat ukur', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 4, code: 'BR-01', name: 'Bor Listrik Bosch', category: 'Alat listrik', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 5, code: 'BR-02', name: 'Bor Baterai Makita', category: 'Alat listrik', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 6, code: 'JS-01', name: 'Jangka Sorong 150mm', category: 'Alat ukur', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 7, code: 'JS-02', name: 'Jangka Sorong Digital Mitutoyo', category: 'Alat ukur', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 8, code: 'JS-03', name: 'Jangka Sorong 300mm', category: 'Alat ukur', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 9, code: 'OL-01', name: 'Obeng Listrik Presisi', category: 'Alat listrik', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 10, code: 'OL-02', name: 'Obeng Torsi Listrik', category: 'Alat listrik', max_hours: 24, condition: 'rusak', condition_note: 'Motor tidak menyala & bau sangit', active: 1 },
  { id: 11, code: 'KM-01', name: 'Kunci Momen Tohnichi', category: 'Kunci khusus', max_hours: 24, condition: 'baik', condition_note: '', active: 1 },
  { id: 12, code: 'KS-01', name: 'Kunci Sok Set 24 pcs', category: 'Kunci khusus', max_hours: 24, condition: 'baik', condition_note: '', active: 1 }
];

// Seed loans (with overdue and recent ones)
const now = new Date();
const twoDaysAgo = new Date(now.getTime() - 48 * 3600 * 1000);
const oneDayAgo = new Date(now.getTime() - 28 * 3600 * 1000);
const twoHoursAgo = new Date(now.getTime() - 2.5 * 3600 * 1000);
const fourHoursAgo = new Date(now.getTime() - 4.5 * 3600 * 1000);

const DEFAULT_LOANS = [
  { id: 101, tool_id: 2, technician_id: 1, borrowed_at: twoDaysAgo.toISOString(), borrow_shift: 'pagi', returned_at: null, return_condition: null, return_note: '', voided_at: null },
  { id: 102, tool_id: 4, technician_id: 2, borrowed_at: oneDayAgo.toISOString(), borrow_shift: 'malam', returned_at: null, return_condition: null, return_note: '', voided_at: null },
  { id: 103, tool_id: 7, technician_id: 4, borrowed_at: twoHoursAgo.toISOString(), borrow_shift: 'pagi', returned_at: null, return_condition: null, return_note: '', voided_at: null },
  { id: 104, tool_id: 8, technician_id: 5, borrowed_at: fourHoursAgo.toISOString(), borrow_shift: 'pagi', returned_at: null, return_condition: null, return_note: '', voided_at: null },
  { id: 105, tool_id: 10, technician_id: 8, borrowed_at: twoDaysAgo.toISOString(), borrow_shift: 'pagi', returned_at: oneDayAgo.toISOString(), return_shift: 'malam', return_condition: 'rusak', return_note: 'Motor tidak menyala & bau sangit', voided_at: null }
];

// App State
let db = {
  technicians: JSON.parse(localStorage.getItem('ag_techs')) || DEFAULT_TECHNICIANS,
  tools: JSON.parse(localStorage.getItem('ag_tools')) || DEFAULT_TOOLS,
  loans: JSON.parse(localStorage.getItem('ag_loans')) || DEFAULT_LOANS,
  settings: JSON.parse(localStorage.getItem('ag_settings')) || {
    default_max_hours: 24,
    shift_pagi_start: '07:00',
    shift_malam_start: '19:00'
  }
};

let lastAction = null; // For undo support

function saveDB() {
  localStorage.setItem('ag_techs', JSON.stringify(db.technicians));
  localStorage.setItem('ag_tools', JSON.stringify(db.tools));
  localStorage.setItem('ag_loans', JSON.stringify(db.loans));
  localStorage.setItem('ag_settings', JSON.stringify(db.settings));
}

// Helpers
function formatRel(isoStr) {
  if (!isoStr) return '';
  const d = new Date(isoStr);
  const diffHours = Math.round((new Date() - d) / (1000 * 3600));
  if (diffHours < 1) return 'Kurang dari 1 jam lalu';
  if (diffHours < 24) return `${diffHours} jam lalu`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays} hari lalu`;
}

function formatClock(isoStr) {
  if (!isoStr) return '';
  const d = new Date(isoStr);
  const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
  const dayName = days[d.getDay()];
  const h = String(d.getHours()).padStart(2, '0');
  const m = String(d.getMinutes()).padStart(2, '0');
  return `${dayName} ${h}:${m}`;
}

function getShift(d = new Date()) {
  const h = d.getHours();
  return (h >= 7 && h < 19) ? 'pagi' : 'malam';
}

function getToolStatus(tool) {
  if (tool.condition === 'rusak') return 'rusak';
  const openLoan = db.loans.find(l => l.tool_id === tool.id && !l.returned_at && !l.voided_at);
  if (openLoan) return 'dipinjam';
  return 'tersedia';
}

function getOpenLoan(toolId) {
  return db.loans.find(l => l.tool_id === toolId && !l.returned_at && !l.voided_at);
}

function getLastLoan(toolId) {
  const list = db.loans.filter(l => l.tool_id === toolId && !l.voided_at);
  list.sort((a, b) => new Date(b.borrowed_at) - new Date(a.borrowed_at));
  return list[0] || null;
}

// Navigation View Switcher
function switchView(viewName, subParam) {
  document.querySelectorAll('main > div').forEach(div => div.classList.add('hidden'));
  document.querySelectorAll('.nav-tab').forEach(b => {
    b.classList.remove('bg-slate-900', 'text-white', 'shadow-sm');
    b.classList.add('text-slate-700');
  });

  const targetView = document.getElementById(`view-${viewName}`);
  if (targetView) targetView.classList.remove('hidden');

  const navBtn = document.getElementById(`nav-${viewName}`);
  if (navBtn) {
    navBtn.classList.add('bg-slate-900', 'text-white', 'shadow-sm');
    navBtn.classList.remove('text-slate-700');
  }

  if (viewName === 'beranda') renderBeranda();
  if (viewName === 'pinjam') renderPinjamForm();
  if (viewName === 'alat') renderSemuaAlat();
  if (viewName === 'kelola') {
    switchKelolaTab(subParam || 'alat');
  }
}

// ===================== BERANDA VIEW =====================
function renderBeranda() {
  const q = (document.getElementById('search-beranda')?.value || '').toLowerCase().trim();

  // Find all active loans
  const openLoans = db.loans.filter(l => !l.returned_at && !l.voided_at);
  
  // Sort oldest first
  openLoans.sort((a, b) => new Date(a.borrowed_at) - new Date(b.borrowed_at));

  const overdueList = [];
  const normalList = [];

  openLoans.forEach(loan => {
    const tool = db.tools.find(t => t.id === loan.tool_id);
    const tech = db.technicians.find(t => t.id === loan.technician_id);
    if (!tool || !tech) return;

    // Filter check
    if (q) {
      const match = tool.name.toLowerCase().includes(q) ||
                    tool.code.toLowerCase().includes(q) ||
                    tech.name.toLowerCase().includes(q) ||
                    tech.section.toLowerCase().includes(q);
      if (!match) return;
    }

    const maxHours = tool.max_hours || db.settings.default_max_hours || 24;
    const hoursElapsed = (new Date() - new Date(loan.borrowed_at)) / (3600 * 1000);
    const item = { loan, tool, tech, hoursElapsed };

    if (hoursElapsed > maxHours) {
      overdueList.push(item);
    } else {
      normalList.push(item);
    }
  });

  // Damaged tools
  const damagedTools = db.tools.filter(t => t.condition === 'rusak' && t.active);
  const damagedFiltered = damagedTools.filter(t => {
    if (!q) return true;
    return t.name.toLowerCase().includes(q) || t.code.toLowerCase().includes(q);
  });

  // Update Stat Cards
  document.getElementById('stat-dipinjam').textContent = openLoans.length;
  document.getElementById('stat-perlu-dicek').textContent = overdueList.length;
  document.getElementById('stat-rusak').textContent = damagedTools.length;

  const cardPerluDicek = document.getElementById('card-perlu-dicek');
  if (overdueList.length > 0) {
    cardPerluDicek.classList.add('border-amber-400', 'bg-amber-50/40', 'ring-4', 'ring-amber-100');
    cardPerluDicek.classList.remove('border-slate-300');
  } else {
    cardPerluDicek.classList.remove('border-amber-400', 'bg-amber-50/40', 'ring-4', 'ring-amber-100');
    cardPerluDicek.classList.add('border-slate-300');
  }

  // Render Perlu Dicek Section
  const secPerluDicek = document.getElementById('section-perlu-dicek');
  const listPerluDicek = document.getElementById('list-perlu-dicek');
  document.getElementById('badge-perlu-dicek-count').textContent = `${overdueList.length} Alat`;

  if (overdueList.length === 0) {
    secPerluDicek.classList.add('hidden');
    listPerluDicek.innerHTML = '';
  } else {
    secPerluDicek.classList.remove('hidden');
    listPerluDicek.innerHTML = overdueList.map(item => `
      <div class="bg-amber-50 border-2 border-amber-300 rounded-2xl p-5 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div class="flex items-start gap-4">
          <span class="badge-warn self-start shrink-0 mt-0.5">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
            Perlu dicek
          </span>
          <div>
            <div class="flex items-center gap-3 flex-wrap">
              <span class="font-mono text-base font-black px-2.5 py-1 bg-amber-200 text-amber-950 rounded-lg">${item.tool.code}</span>
              <button onclick="showToolDetail(${item.tool.id})" class="text-2xl font-black text-slate-900 hover:text-amber-700 underline-offset-4 hover:underline text-left">
                ${item.tool.name}
              </button>
            </div>
            <div class="text-lg text-slate-800 font-semibold mt-1">
              Peminjam: <strong class="text-slate-950 text-xl">${item.tech.name}</strong> 
              <span class="text-slate-500 font-medium">(${item.tech.section})</span>
            </div>
            <div class="text-base text-amber-900 font-bold mt-0.5 flex items-center gap-2">
              <svg class="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
              Sejak ${formatClock(item.loan.borrowed_at)} &bull; <span class="underline decoration-amber-400 font-black">${formatRel(item.loan.borrowed_at)}</span>
            </div>
          </div>
        </div>
        <button onclick="openModalKembali(${item.loan.id})" 
                class="btn-warn text-slate-950 font-black min-h-14 px-7 w-full md:w-auto shadow-md hover:scale-105 active:scale-95 transition-all text-lg shrink-0">
          <svg class="w-6 h-6 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          Dikembalikan
        </button>
      </div>
    `).join('');
  }

  // Render Sedang Dipinjam Section
  const listDipinjam = document.getElementById('list-dipinjam');
  document.getElementById('badge-dipinjam-count').textContent = `${normalList.length} Alat`;

  if (normalList.length === 0) {
    listDipinjam.innerHTML = `
      <div class="bg-white border-2 border-dashed border-slate-300 rounded-2xl p-8 text-center">
        <p class="text-xl font-bold text-slate-600">Tidak ada alat aktif yang sedang dipinjam.</p>
        <p class="text-base text-slate-400 mt-1">Gudang sparepart dalam kondisi aman.</p>
      </div>
    `;
  } else {
    listDipinjam.innerHTML = normalList.map(item => `
      <div class="bg-white border-2 border-slate-300 rounded-2xl p-5 shadow-sm hover:border-slate-400 transition-all flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div class="flex items-start gap-4">
          <span class="badge-info self-start shrink-0 mt-0.5">
            Dipinjam
          </span>
          <div>
            <div class="flex items-center gap-3 flex-wrap">
              <span class="font-mono text-base font-black px-2.5 py-1 bg-slate-200 text-slate-800 rounded-lg">${item.tool.code}</span>
              <button onclick="showToolDetail(${item.tool.id})" class="text-2xl font-black text-slate-900 hover:text-emerald-700 underline-offset-4 hover:underline text-left">
                ${item.tool.name}
              </button>
            </div>
            <div class="text-lg text-slate-800 font-semibold mt-1">
              Peminjam: <strong class="text-slate-950 text-xl">${item.tech.name}</strong> 
              <span class="text-slate-500 font-medium">(${item.tech.section})</span>
            </div>
            <div class="text-base text-slate-600 font-medium mt-0.5 flex items-center gap-2">
              <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
              Sejak ${formatClock(item.loan.borrowed_at)} (${formatRel(item.loan.borrowed_at)})
            </div>
          </div>
        </div>
        <button onclick="openModalKembali(${item.loan.id})" 
                class="btn-primary min-h-14 px-7 w-full md:w-auto shadow-md hover:scale-105 active:scale-95 transition-all text-lg shrink-0">
          <svg class="w-6 h-6 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          Dikembalikan
        </button>
      </div>
    `).join('');
  }

  // Render Rusak Section
  const secRusak = document.getElementById('section-rusak');
  const listRusak = document.getElementById('list-rusak');
  document.getElementById('badge-rusak-count').textContent = `${damagedFiltered.length} Alat`;

  if (damagedFiltered.length === 0) {
    secRusak.classList.add('hidden');
    listRusak.innerHTML = '';
  } else {
    secRusak.classList.remove('hidden');
    listRusak.innerHTML = damagedFiltered.map(tool => {
      const lastLoan = getLastLoan(tool.id);
      const borrowerName = lastLoan ? (db.technicians.find(t => t.id === lastLoan.technician_id)?.name || '-') : '-';
      const returnTime = lastLoan && lastLoan.returned_at ? formatClock(lastLoan.returned_at) : '';

      return `
        <div class="bg-rose-50/70 border-2 border-rose-300 rounded-2xl p-5 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
          <div class="flex items-start gap-4">
            <span class="badge-danger self-start shrink-0 mt-0.5">
              Rusak
            </span>
            <div>
              <div class="flex items-center gap-3 flex-wrap">
                <span class="font-mono text-base font-black px-2.5 py-1 bg-rose-200 text-rose-950 rounded-lg">${tool.code}</span>
                <button onclick="showToolDetail(${tool.id})" class="text-2xl font-black text-slate-900 hover:text-rose-700 underline-offset-4 hover:underline text-left">
                  ${tool.name}
                </button>
              </div>
              <div class="text-base text-slate-700 font-medium mt-1">
                Peminjam terakhir: <strong class="text-slate-900 font-bold">${borrowerName}</strong>
                ${returnTime ? `&bull; Dikembalikan ${returnTime}` : ''}
              </div>
              ${tool.condition_note ? `
                <div class="text-base text-rose-950 font-bold mt-1 bg-rose-100/90 px-3 py-1 rounded-lg inline-block border border-rose-300">
                  Catatan: ${tool.condition_note}
                </div>
              ` : ''}
            </div>
          </div>
          <button onclick="tandaiSudahDiperbaiki(${tool.id})" 
                  class="btn-secondary min-h-14 px-6 border-slate-400 hover:bg-emerald-50 hover:border-emerald-500 hover:text-emerald-900 text-slate-800 font-bold text-lg w-full md:w-auto shadow-sm transition-all flex items-center justify-center gap-2 shrink-0">
            <svg class="w-6 h-6 text-emerald-600 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path></svg>
            Sudah Diperbaiki
          </button>
        </div>
      `;
    }).join('');
  }
}

// ===================== PINJAM FORM VIEW =====================
let selectedTechId = null;
let selectedToolIds = new Set();

function renderPinjamForm() {
  selectedTechId = null;
  selectedToolIds.clear();
  updatePinjamSummary();

  // Render Technicians (Step 1)
  const techGrid = document.getElementById('pinjam-tech-grid');
  techGrid.innerHTML = db.technicians.filter(t => t.active).map(tech => `
    <div onclick="pilihTeknisi(${tech.id})" 
         id="tech-card-${tech.id}"
         class="tech-card p-4 rounded-2xl border-2 border-slate-300 bg-slate-50 hover:bg-slate-100 hover:border-slate-400 cursor-pointer transition-all text-center select-none">
      <div class="font-black text-xl text-slate-900">${tech.name}</div>
      <div class="text-sm font-bold text-slate-500 mt-0.5">${tech.section}</div>
    </div>
  `).join('');

  // Render Tools (Step 2)
  renderPinjamToolsGrid('all');
}

function renderPinjamToolsGrid(category) {
  const toolsGrid = document.getElementById('pinjam-tools-grid');
  toolsGrid.innerHTML = db.tools.filter(t => t.active).map(tool => {
    if (category !== 'all' && tool.category !== category) return '';

    const status = getToolStatus(tool);
    const isLoanable = (status === 'tersedia');
    const openLoan = getOpenLoan(tool.id);
    const borrowerName = openLoan ? (db.technicians.find(t => t.id === openLoan.technician_id)?.name || 'teknisi') : '';

    return `
      <div onclick="${isLoanable ? `togglePilihAlat(${tool.id})` : ''}"
           id="pinjam-tool-card-${tool.id}"
           class="p-4 rounded-2xl border-2 transition-all select-none ${
             isLoanable 
               ? 'border-slate-300 bg-white hover:bg-slate-50 hover:border-slate-400 cursor-pointer' 
               : 'border-slate-200 bg-slate-100/80 cursor-not-allowed opacity-60'
           }">
        <div class="flex items-start justify-between gap-2">
          <span class="font-mono text-sm font-black px-2 py-0.5 rounded ${isLoanable ? 'bg-slate-100 text-slate-800' : 'bg-slate-200 text-slate-500'}">
            ${tool.code}
          </span>
          ${
            tool.condition === 'rusak' 
              ? '<span class="text-xs font-black px-2 py-0.5 bg-rose-100 text-rose-800 rounded">RUSAK</span>' 
              : status === 'dipinjam' 
                ? '<span class="text-xs font-black px-2 py-0.5 bg-amber-100 text-amber-900 rounded">DIPINJAM</span>' 
                : '<span class="text-xs font-black px-2 py-0.5 bg-emerald-100 text-emerald-900 rounded">TERSEDIA</span>'
          }
        </div>
        <div class="font-black text-xl text-slate-900 mt-2 leading-snug">${tool.name}</div>
        <div class="text-xs font-bold ${isLoanable ? 'text-slate-400 mt-2' : 'text-slate-500 mt-2'}">
          ${
            tool.condition === 'rusak'
              ? 'Tidak bisa dipinjam (sedang rusak)'
              : status === 'dipinjam'
                ? `Sedang dibawa: <strong>${borrowerName}</strong>`
                : 'Klik untuk memilih'
          }
        </div>
      </div>
    `;
  }).join('');
}

function filterPinjamCat(cat, btn) {
  document.querySelectorAll('#pinjam-cat-filters .pinjam-cat-btn').forEach(b => {
    b.classList.remove('bg-slate-900', 'text-white');
    b.classList.add('bg-slate-100', 'text-slate-700');
  });
  btn.classList.add('bg-slate-900', 'text-white');
  btn.classList.remove('bg-slate-100', 'text-slate-700');
  renderPinjamToolsGrid(cat);

  // Re-apply selected styles
  selectedToolIds.forEach(id => {
    const card = document.getElementById(`pinjam-tool-card-${id}`);
    if (card) {
      card.classList.add('border-emerald-600', 'bg-emerald-50/70', 'ring-2', 'ring-emerald-500');
    }
  });
}

function pilihTeknisi(id) {
  selectedTechId = id;
  document.querySelectorAll('.tech-card').forEach(c => {
    c.classList.remove('border-emerald-600', 'bg-emerald-50', 'ring-4', 'ring-emerald-300');
    c.classList.add('border-slate-300', 'bg-slate-50');
  });
  const el = document.getElementById(`tech-card-${id}`);
  if (el) {
    el.classList.add('border-emerald-600', 'bg-emerald-50', 'ring-4', 'ring-emerald-300');
    el.classList.remove('border-slate-300', 'bg-slate-50');
  }
  updatePinjamSummary();
}

function togglePilihAlat(id) {
  const card = document.getElementById(`pinjam-tool-card-${id}`);
  if (selectedToolIds.has(id)) {
    selectedToolIds.delete(id);
    if (card) {
      card.classList.remove('border-emerald-600', 'bg-emerald-50/70', 'ring-2', 'ring-emerald-500');
      card.classList.add('border-slate-300', 'bg-white');
    }
  } else {
    selectedToolIds.add(id);
    if (card) {
      card.classList.add('border-emerald-600', 'bg-emerald-50/70', 'ring-2', 'ring-emerald-500');
      card.classList.remove('border-slate-300', 'bg-white');
    }
  }
  updatePinjamSummary();
}

function updatePinjamSummary() {
  const btn = document.getElementById('btn-confirm-pinjam');
  const summaryText = document.getElementById('pinjam-summary-text');
  const summaryHint = document.getElementById('pinjam-summary-hint');

  const tech = db.technicians.find(t => t.id === selectedTechId);
  const selectedTools = Array.from(selectedToolIds).map(id => db.tools.find(t => t.id === id)).filter(Boolean);

  if (tech && selectedTools.length > 0) {
    btn.removeAttribute('disabled');
    const names = selectedTools.map(t => `${t.name} (${t.code})`).join(', ');
    summaryText.innerHTML = `<span class="text-emerald-400">${tech.name}</span> meminjam <strong class="underline">${selectedTools.length} alat</strong>: ${names}`;
    summaryHint.textContent = 'Klik tombol hijau di sebelah kanan untuk menyimpan transaksi.';
  } else if (tech && selectedTools.length === 0) {
    btn.setAttribute('disabled', 'true');
    summaryText.textContent = `${tech.name} meminjam... (sekarang pilih alatnya di atas)`;
    summaryHint.textContent = 'Centang minimal satu alat yang berstatus "Tersedia".';
  } else {
    btn.setAttribute('disabled', 'true');
    summaryText.textContent = 'Belum ada teknisi dan alat yang dipilih.';
    summaryHint.textContent = 'Pilih 1 nama teknisi di atas, lalu centang minimal 1 alat.';
  }
}

function submitPeminjaman() {
  if (!selectedTechId || selectedToolIds.size === 0) return;

  const tech = db.technicians.find(t => t.id === selectedTechId);
  const toolIds = Array.from(selectedToolIds);
  const shift = getShift();
  const borrowTime = new Date().toISOString();

  const newLoanIds = [];
  toolIds.forEach(toolId => {
    const newId = Date.now() + Math.floor(Math.random() * 1000);
    db.loans.push({
      id: newId,
      tool_id: toolId,
      technician_id: selectedTechId,
      borrowed_at: borrowTime,
      borrow_shift: shift,
      returned_at: null,
      return_condition: null,
      return_note: '',
      voided_at: null
    });
    newLoanIds.push(newId);
  });

  saveDB();

  // Register undo action
  lastAction = {
    type: 'borrow',
    loanIds: newLoanIds,
    techName: tech.name,
    count: toolIds.length
  };

  showToast({
    kind: 'success',
    message: `Tersimpan: ${tech.name} meminjam ${toolIds.length} alat.`,
    subtext: `Shift ${shift.toUpperCase()} &bull; Dicatat otomatis`,
    canUndo: true,
    action: lastAction
  });

  switchView('beranda');
}

// ===================== RETURN (DIKEMBALIKAN) FLOW =====================
let currentReturnLoanId = null;

function openModalKembali(loanId) {
  currentReturnLoanId = loanId;
  const loan = db.loans.find(l => l.id === loanId);
  if (!loan) return;
  const tool = db.tools.find(t => t.id === loan.tool_id);
  const tech = db.technicians.find(t => t.id === loan.technician_id);

  const container = document.getElementById('modal-container');
  const content = document.getElementById('modal-content');

  content.innerHTML = `
    <!-- Header -->
    <div class="flex items-start justify-between pb-4 border-b border-slate-200">
      <div>
        <span class="text-xs font-bold uppercase tracking-wider text-slate-500 bg-slate-100 px-3 py-1 rounded-md">Konfirmasi Pengembalian</span>
        <h3 class="text-2xl sm:text-3xl font-black text-slate-900 mt-2">${tool.name}</h3>
        <div class="text-lg text-slate-600 font-semibold mt-1">
          Kode: <span class="font-mono text-emerald-800 bg-emerald-50 px-2 py-0.5 rounded">${tool.code}</span> &bull; Peminjam: <strong class="text-slate-900">${tech.name}</strong>
        </div>
      </div>
      <button onclick="closeModal()" class="w-12 h-12 flex items-center justify-center rounded-2xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 text-2xl font-bold cursor-pointer">
        &times;
      </button>
    </div>

    <!-- Step 1: 2 Big Buttons (Baik vs Rusak) -->
    <div id="modal-step-initial" class="py-6 space-y-4">
      <p class="text-lg font-medium text-slate-700">Bagaimana kondisi alat saat dikembalikan?</p>
      
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <!-- 1-Click Fast Return "BAIK" (2 clicks total from dashboard!) -->
        <button onclick="prosesKembali('baik', '')" 
                class="w-full min-h-24 p-5 bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white rounded-2xl shadow-md flex flex-col items-center justify-center text-center cursor-pointer transition-all hover:scale-[1.02] focus:ring-4 focus:ring-emerald-300">
          <svg class="w-10 h-10 mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path></svg>
          <span class="text-2xl font-black tracking-wide">KONDISI BAIK</span>
          <span class="text-xs text-emerald-100 font-medium mt-1">Siap dipinjamkan kembali</span>
        </button>

        <!-- Option "ADA KERUSAKAN" -->
        <button onclick="showDamageForm()" 
                class="w-full min-h-24 p-5 bg-rose-50 hover:bg-rose-100 active:bg-rose-200 text-rose-800 border-2 border-rose-300 rounded-2xl shadow-sm flex flex-col items-center justify-center text-center cursor-pointer transition-all hover:scale-[1.02] focus:ring-4 focus:ring-rose-200">
          <svg class="w-10 h-10 mb-1 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
          <span class="text-2xl font-black tracking-wide">ADA KERUSAKAN</span>
          <span class="text-xs text-rose-600 font-medium mt-1">Perlu perbaikan gudang</span>
        </button>
      </div>
    </div>

    <!-- Step 2: Damage details form -->
    <div id="modal-step-damage" class="hidden py-4 space-y-5">
      <div class="bg-rose-50 border border-rose-200 p-4 rounded-xl flex items-center gap-3 text-rose-900">
        <svg class="w-6 h-6 shrink-0 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        <span class="text-sm font-semibold">Alat akan ditandai <strong class="font-bold underline">RUSAK</strong> dan tidak dapat dipinjam sampai diperbaiki.</span>
      </div>

      <div>
        <label class="block text-base font-bold text-slate-800 mb-2">Pilih Jenis Kerusakan (Klik salah satu):</label>
        <div class="grid grid-cols-2 gap-2.5" id="modal-damage-chips">
          <button type="button" onclick="selectChip(this, 'Tidak menyala')" class="chip-option">🔌 Tidak menyala</button>
          <button type="button" onclick="selectChip(this, 'Patah/aus')" class="chip-option">🔨 Patah / aus</button>
          <button type="button" onclick="selectChip(this, 'Aksesori hilang')" class="chip-option">📦 Aksesori hilang</button>
          <button type="button" onclick="selectChip(this, 'Lainnya')" class="chip-option">⚠️ Lainnya</button>
        </div>
      </div>

      <div>
        <label class="block text-base font-bold text-slate-800 mb-1">Catatan Tambahan (Opsional):</label>
        <input type="text" id="modal-catatan-input" placeholder="Contoh: Kabel aus dekat sambungan" class="w-full text-lg px-4 py-3 border-2 border-slate-300 rounded-xl focus:border-rose-500 focus:ring-4 focus:ring-rose-200 outline-none">
      </div>

      <div class="flex items-center justify-end gap-3 pt-4 border-t border-slate-200">
        <button type="button" onclick="hideDamageForm()" class="btn-secondary min-h-12 py-2 px-5 text-base">Kembali</button>
        <button type="button" onclick="submitKerusakan()" class="btn-danger min-h-12 py-2 px-6 text-lg">Simpan Kerusakan</button>
      </div>
    </div>

    <!-- Footer Cancel -->
    <div class="pt-4 border-t border-slate-100 flex justify-end">
      <button type="button" onclick="closeModal()" class="px-5 py-2.5 text-slate-600 hover:text-slate-900 font-bold text-base hover:bg-slate-100 rounded-xl transition-colors cursor-pointer">
        Batal (Jangan Ubah)
      </button>
    </div>
  `;

  container.classList.remove('hidden');
}

function closeModal() {
  document.getElementById('modal-container').classList.add('hidden');
  currentReturnLoanId = null;
}

function showDamageForm() {
  document.getElementById('modal-step-initial').classList.add('hidden');
  document.getElementById('modal-step-damage').classList.remove('hidden');
}

function hideDamageForm() {
  document.getElementById('modal-step-damage').classList.add('hidden');
  document.getElementById('modal-step-initial').classList.remove('hidden');
}

let selectedChipValue = '';
function selectChip(btn, val) {
  document.querySelectorAll('#modal-damage-chips .chip-option').forEach(el => el.classList.remove('selected'));
  btn.classList.add('selected');
  selectedChipValue = val;
  const input = document.getElementById('modal-catatan-input');
  if (input && !input.value) input.value = val;
}

function submitKerusakan() {
  const note = document.getElementById('modal-catatan-input').value.trim() || selectedChipValue || 'Ada kerusakan';
  prosesKembali('rusak', note);
}

function prosesKembali(condition, note) {
  const loan = db.loans.find(l => l.id === currentReturnLoanId);
  if (!loan) return;
  const tool = db.tools.find(t => t.id === loan.tool_id);
  const tech = db.technicians.find(t => t.id === loan.technician_id);

  loan.returned_at = new Date().toISOString();
  loan.return_shift = getShift();
  loan.return_condition = condition;
  loan.return_note = note;

  if (condition === 'rusak') {
    tool.condition = 'rusak';
    tool.condition_note = note;
  }

  saveDB();
  closeModal();

  lastAction = {
    type: 'return',
    loanId: loan.id,
    toolId: tool.id,
    prevCondition: 'baik',
    toolName: tool.name,
    techName: tech.name
  };

  showToast({
    kind: condition === 'baik' ? 'success' : 'warn',
    message: `${tool.name} telah dikembalikan oleh ${tech.name} (${condition === 'baik' ? 'Kondisi Baik' : 'Ditandai Rusak'}).`,
    subtext: `Shift ${loan.return_shift.toUpperCase()} &bull; Peminjaman selesai`,
    canUndo: true,
    action: lastAction
  });

  renderBeranda();
}

// ===================== SUDAH DIPERBAIKI ACTION =====================
function tandaiSudahDiperbaiki(toolId) {
  const tool = db.tools.find(t => t.id === toolId);
  if (!tool) return;
  tool.condition = 'baik';
  tool.condition_note = '';
  saveDB();

  showToast({
    kind: 'success',
    message: `${tool.name} telah ditandai sudah diperbaiki.`,
    subtext: 'Status alat kembali Tersedia untuk dipinjamkan.',
    canUndo: false
  });

  renderBeranda();
  renderSemuaAlat();
}

// ===================== UNDO (BATALKAN) HANDLER =====================
function triggerUndo() {
  if (!lastAction) return;

  if (lastAction.type === 'borrow') {
    // Void loans
    lastAction.loanIds.forEach(id => {
      const loan = db.loans.find(l => l.id === id);
      if (loan) loan.voided_at = new Date().toISOString();
    });
    saveDB();
    showToast({
      kind: 'warn',
      message: `Peminjaman ${lastAction.techName} (${lastAction.count} alat) berhasil dibatalkan.`,
      canUndo: false
    });
  } else if (lastAction.type === 'return') {
    const loan = db.loans.find(l => l.id === lastAction.loanId);
    const tool = db.tools.find(t => t.id === lastAction.toolId);
    if (loan) {
      loan.returned_at = null;
      loan.return_shift = null;
      loan.return_condition = null;
      loan.return_note = '';
    }
    if (tool && tool.condition === 'rusak') {
      tool.condition = 'baik';
      tool.condition_note = '';
    }
    saveDB();
    showToast({
      kind: 'warn',
      message: `Pengembalian ${lastAction.toolName} berhasil dibatalkan (kembali berstatus dipinjam).`,
      canUndo: false
    });
  }

  lastAction = null;
  renderBeranda();
  renderSemuaAlat();
}

// ===================== TOAST NOTIFICATION =====================
function showToast({ kind, message, subtext, canUndo, action }) {
  const container = document.getElementById('toast-container');
  container.innerHTML = `
    <div class="rounded-2xl p-5 mb-4 shadow-lg border-2 flex items-center justify-between gap-4 transition-all ${
      kind === 'success' 
        ? 'bg-emerald-50 border-emerald-500 text-emerald-950' 
        : kind === 'warn' 
          ? 'bg-amber-50 border-amber-500 text-amber-950' 
          : 'bg-rose-50 border-rose-500 text-rose-950'
    }">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl flex items-center justify-center text-white shrink-0 ${
          kind === 'success' ? 'bg-emerald-600' : kind === 'warn' ? 'bg-amber-600' : 'bg-rose-600'
        }">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path></svg>
        </div>
        <div>
          <div class="text-xl font-bold leading-tight">${message}</div>
          ${subtext ? `<div class="text-base text-slate-600 font-medium mt-0.5">${subtext}</div>` : ''}
        </div>
      </div>
      <div class="flex items-center gap-3 shrink-0">
        ${canUndo ? `
          <button onclick="triggerUndo()" class="px-5 py-2.5 bg-slate-900 hover:bg-slate-800 text-white font-black text-base rounded-xl shadow transition-colors flex items-center gap-2 cursor-pointer focus:ring-4 focus:ring-slate-400">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"></path></svg>
            Batalkan
          </button>
        ` : ''}
        <button onclick="document.getElementById('toast-container').innerHTML = ''" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-200/60 rounded-xl cursor-pointer">
          &times;
        </button>
      </div>
    </div>
  `;
}

// ===================== SEMUA ALAT VIEW =====================
let filterSemuaStatus = 'all';

function filterSemuaAlatStatus(status, btn) {
  document.querySelectorAll('#semua-alat-status-filters .sat-btn').forEach(b => {
    b.classList.remove('bg-slate-900', 'text-white');
    b.classList.add('bg-slate-100', 'text-slate-700');
  });
  btn.classList.add('bg-slate-900', 'text-white');
  btn.classList.remove('bg-slate-100', 'text-slate-700');
  filterSemuaStatus = status;
  renderSemuaAlat();
}

function renderSemuaAlat() {
  const tbody = document.getElementById('semua-alat-tbody');
  const q = (document.getElementById('search-semua-alat')?.value || '').toLowerCase().trim();

  const filtered = db.tools.filter(tool => {
    if (!tool.active) return false;
    const status = getToolStatus(tool);
    if (filterSemuaStatus !== 'all' && status !== filterSemuaStatus) return false;
    if (q) {
      return tool.name.toLowerCase().includes(q) ||
             tool.code.toLowerCase().includes(q) ||
             tool.category.toLowerCase().includes(q);
    }
    return true;
  });

  tbody.innerHTML = filtered.map(tool => {
    const status = getToolStatus(tool);
    const openLoan = getOpenLoan(tool.id);
    const borrower = openLoan ? (db.technicians.find(t => t.id === openLoan.technician_id)?.name || '') : '';
    const lastLoan = getLastLoan(tool.id);
    const lastBorrower = lastLoan ? (db.technicians.find(t => t.id === lastLoan.technician_id)?.name || '') : '';

    return `
      <tr class="hover:bg-slate-50/80 transition-colors">
        <td class="py-4 px-6">
          <span class="font-mono text-base font-black px-2.5 py-1 bg-slate-100 border border-slate-300 text-slate-900 rounded-lg">
            ${tool.code}
          </span>
        </td>
        <td class="py-4 px-6">
          <button onclick="showToolDetail(${tool.id})" class="font-black text-xl text-slate-900 hover:text-emerald-700 hover:underline text-left">
            ${tool.name}
          </button>
          ${tool.condition_note ? `<div class="text-xs text-rose-700 font-bold mt-0.5">Catatan: ${tool.condition_note}</div>` : ''}
        </td>
        <td class="py-4 px-6 text-slate-600 font-semibold">${tool.category || '-'}</td>
        <td class="py-4 px-6 text-slate-600">${tool.max_hours || 24} Jam</td>
        <td class="py-4 px-6">
          ${
            status === 'rusak' 
              ? `<span class="badge-danger">Rusak</span>${lastBorrower ? `<div class="text-xs text-slate-600 mt-1 font-semibold">Peminjam terakhir: ${lastBorrower}</div>` : ''}`
              : status === 'dipinjam'
                ? `<span class="badge-info">Dipinjam</span><div class="text-xs text-slate-900 mt-1 font-bold">Oleh: ${borrower}</div>`
                : `<span class="badge-success">Tersedia</span>`
          }
        </td>
        <td class="py-4 px-6 text-right space-x-2">
          ${
            status === 'tersedia'
              ? `
                <button onclick="pinjamkanSatuAlat(${tool.id})" class="inline-flex items-center gap-1.5 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white font-bold text-sm rounded-xl shadow-xs transition-colors cursor-pointer">
                  <svg class="w-4 h-4 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"></path></svg>
                  Pinjamkan
                </button>
              `
              : status === 'rusak'
                ? `
                  <button onclick="tandaiSudahDiperbaiki(${tool.id})" class="px-4 py-2 bg-slate-800 hover:bg-slate-900 text-white font-bold text-sm rounded-xl transition-colors cursor-pointer">
                    Sudah Diperbaiki
                  </button>
                `
                : ''
          }
          <button onclick="showToolDetail(${tool.id})" class="inline-flex items-center gap-1 px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-sm rounded-xl transition-colors cursor-pointer">
            Riwayat
          </button>
        </td>
      </tr>
    `;
  }).join('');
}

function pinjamkanSatuAlat(toolId) {
  switchView('pinjam');
  togglePilihAlat(toolId);
}

// ===================== DETAIL ALAT VIEW =====================
function showToolDetail(toolId) {
  const tool = db.tools.find(t => t.id === toolId);
  if (!tool) return;
  const status = getToolStatus(tool);
  const lastLoan = getLastLoan(tool.id);
  const lastTech = lastLoan ? db.technicians.find(t => t.id === lastLoan.technician_id) : null;

  // History list
  const history = db.loans.filter(l => l.tool_id === tool.id && !l.voided_at);
  history.sort((a, b) => new Date(b.borrowed_at) - new Date(a.borrowed_at));

  const container = document.getElementById('detail-card-container');
  container.innerHTML = `
    <!-- Tool Overview Card -->
    <div class="bg-white rounded-3xl p-6 sm:p-8 border-2 border-slate-300 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
      <div class="space-y-2">
        <div class="flex items-center gap-3 flex-wrap">
          <span class="font-mono text-xl font-black px-3 py-1 bg-slate-900 text-white rounded-xl">${tool.code}</span>
          <h1 class="text-3xl sm:text-4xl font-black text-slate-900">${tool.name}</h1>
          ${
            status === 'rusak'
              ? '<span class="badge-danger text-base px-3 py-1">Rusak</span>'
              : status === 'dipinjam'
                ? '<span class="badge-info text-base px-3 py-1">Sedang Dipinjam</span>'
                : '<span class="badge-success text-base px-3 py-1">Tersedia di Gudang</span>'
          }
        </div>
        <div class="flex items-center gap-6 text-base text-slate-600 font-medium pt-1">
          <span>Kategori: <strong class="text-slate-800">${tool.category || '-'}</strong></span>
          <span>Batas Waktu: <strong class="text-slate-800">${tool.max_hours || 24} Jam</strong></span>
        </div>
        ${tool.condition_note ? `
          <div class="bg-rose-50 border border-rose-200 text-rose-900 px-4 py-2.5 rounded-xl font-semibold text-base mt-2">
            Catatan Kerusakan: ${tool.condition_note}
          </div>
        ` : ''}
      </div>

      <div class="shrink-0 flex items-center gap-3 w-full md:w-auto">
        ${
          status === 'tersedia'
            ? `<button onclick="pinjamkanSatuAlat(${tool.id})" class="btn-primary min-h-14 px-8 w-full md:w-auto">＋ Pinjamkan Alat Ini</button>`
            : status === 'rusak'
              ? `<button onclick="tandaiSudahDiperbaiki(${tool.id})" class="btn-primary min-h-14 px-8 w-full md:w-auto">Sudah Diperbaiki</button>`
              : ''
        }
      </div>
    </div>

    <!-- Highlight Box: Peminjam Terakhir -->
    <div class="bg-gradient-to-r from-slate-900 to-slate-800 text-white rounded-3xl p-6 sm:p-8 shadow-md border-2 border-slate-700 mt-6">
      <div class="text-xs uppercase font-extrabold tracking-wider text-emerald-400 mb-1">Informasi Peminjam Terakhir:</div>
      ${
        lastLoan && lastTech ? `
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mt-2">
            <div>
              <div class="text-3xl font-black">${lastTech.name}</div>
              <div class="text-base text-slate-300 font-medium mt-1">
                Bagian: <strong class="text-white">${lastTech.section}</strong> &bull;
                Shift: <span class="uppercase font-bold">${lastLoan.borrow_shift}</span>
              </div>
              <div class="text-sm text-slate-400 mt-1">
                Dipinjam: ${formatClock(lastLoan.borrowed_at)} &bull; 
                ${lastLoan.returned_at ? `Dikembalikan: ${formatClock(lastLoan.returned_at)}` : '<strong class="text-amber-400">Belum Dikembalikan</strong>'}
              </div>
            </div>
            ${
              lastLoan.return_condition ? `
                <div class="sm:text-right">
                  <span class="text-xs uppercase font-bold text-slate-400 block mb-1">Kondisi saat kembali</span>
                  ${
                    lastLoan.return_condition === 'baik'
                      ? '<span class="badge-success text-base px-3 py-1.5">Baik</span>'
                      : `<span class="badge-danger text-base px-3 py-1.5">Ada Kerusakan</span>
                         ${lastLoan.return_note ? `<div class="text-xs text-rose-300 font-medium mt-1">${lastLoan.return_note}</div>` : ''}`
                  }
                </div>
              ` : ''
            }
          </div>
        ` : '<div class="text-xl font-bold text-slate-300 mt-2">Belum ada riwayat peminjaman untuk alat ini.</div>'
      }
    </div>

    <!-- Riwayat Peminjaman Tabel -->
    <div class="bg-white rounded-3xl border-2 border-slate-300 shadow-sm p-6 sm:p-8 space-y-4 mt-6">
      <h2 class="text-2xl font-black text-slate-900 uppercase tracking-tight flex items-center justify-between">
        Riwayat Peminjaman
        <span class="text-sm font-semibold text-slate-500 lowercase">catatan permanen &bull; tidak dapat dihapus</span>
      </h2>

      ${
        history.length > 0 ? `
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-slate-50 border-b-2 border-slate-200 text-slate-700 text-sm font-black uppercase">
                  <th class="py-3 px-4">Peminjam</th>
                  <th class="py-3 px-4">Waktu Pinjam</th>
                  <th class="py-3 px-4">Waktu Kembali</th>
                  <th class="py-3 px-4">Kondisi</th>
                  <th class="py-3 px-4">Catatan</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200 text-base">
                ${history.map(item => {
                  const tech = db.technicians.find(t => t.id === item.technician_id);
                  return `
                    <tr class="hover:bg-slate-50">
                      <td class="py-4 px-4 font-bold text-slate-900">
                        ${tech?.name || 'Teknisi'}
                        <div class="text-xs text-slate-500 font-medium">${tech?.section || ''}</div>
                      </td>
                      <td class="py-4 px-4 text-slate-700">
                        ${formatClock(item.borrowed_at)}
                        <span class="text-xs px-2 py-0.5 bg-slate-100 rounded text-slate-600 uppercase font-bold">${item.borrow_shift}</span>
                      </td>
                      <td class="py-4 px-4 text-slate-700">
                        ${
                          item.returned_at
                            ? `${formatClock(item.returned_at)} <span class="text-xs px-2 py-0.5 bg-slate-100 rounded text-slate-600 uppercase font-bold">${item.return_shift || ''}</span>`
                            : '<span class="text-amber-800 font-bold text-sm bg-amber-50 px-2 py-0.5 rounded">Masih Dipinjam</span>'
                        }
                      </td>
                      <td class="py-4 px-4">
                        ${
                          item.return_condition === 'baik' 
                            ? '<span class="badge-success text-xs">Baik</span>' 
                            : item.return_condition === 'rusak' 
                              ? '<span class="badge-danger text-xs">Rusak</span>' 
                              : '-'
                        }
                      </td>
                      <td class="py-4 px-4 text-slate-600 text-sm">${item.return_note || '-'}</td>
                    </tr>
                  `;
                }).join('')}
              </tbody>
            </table>
          </div>
        ` : '<p class="text-slate-500 font-medium py-4 text-center">Belum ada catatan transaksi.</p>'
      }
    </div>
  `;

  switchView('detail');
}

// ===================== KELOLA VIEW =====================
let currentKelolaTab = 'alat';

function switchKelolaTab(tab) {
  currentKelolaTab = tab;
  ['alat', 'teknisi', 'pengaturan'].forEach(t => {
    const el = document.getElementById(`kelola-tab-${t}`);
    const btn = document.getElementById(`tab-btn-${t}`);
    if (t === tab) {
      if (el) el.classList.remove('hidden');
      if (btn) {
        btn.classList.add('bg-slate-900', 'text-white', 'shadow-sm');
        btn.classList.remove('bg-white', 'text-slate-700', 'border');
      }
    } else {
      if (el) el.classList.add('hidden');
      if (btn) {
        btn.classList.remove('bg-slate-900', 'text-white', 'shadow-sm');
        btn.classList.add('bg-white', 'text-slate-700', 'border');
      }
    }
  });

  if (tab === 'alat') renderKelolaAlat();
  if (tab === 'teknisi') renderKelolaTeknisi();
}

function renderKelolaAlat() {
  const tbody = document.getElementById('kelola-alat-tbody');
  tbody.innerHTML = db.tools.map(tool => `
    <tr class="hover:bg-slate-50 ${!tool.active ? 'opacity-50 bg-slate-50' : ''}">
      <td class="py-3 px-3 font-mono font-bold">${tool.code}</td>
      <td class="py-3 px-3 font-bold text-slate-900">${tool.name}</td>
      <td class="py-3 px-3 text-slate-600 text-sm">${tool.category}</td>
      <td class="py-3 px-3">
        ${tool.active ? '<span class="badge-success text-xs">Aktif</span>' : '<span class="badge-warn text-xs">Nonaktif</span>'}
      </td>
      <td class="py-3 px-3 text-right">
        <button onclick="toggleAktifAlat(${tool.id})" class="text-xs font-bold px-3 py-1.5 rounded-lg border border-slate-300 hover:bg-slate-200 transition-colors cursor-pointer">
          ${tool.active ? 'Nonaktifkan' : 'Aktifkan'}
        </button>
      </td>
    </tr>
  `).join('');
}

function tambahAlat(e) {
  e.preventDefault();
  const code = document.getElementById('input-tool-code').value.trim().toUpperCase();
  const name = document.getElementById('input-tool-name').value.trim();
  const cat = document.getElementById('input-tool-cat').value;
  const hours = parseInt(document.getElementById('input-tool-hours').value) || 24;

  if (db.tools.some(t => t.code === code)) {
    alert(`Kode alat ${code} sudah terdaftar! Gunakan kode lain.`);
    return;
  }

  db.tools.push({
    id: Date.now(),
    code,
    name,
    category: cat,
    max_hours: hours,
    condition: 'baik',
    condition_note: '',
    active: 1
  });

  saveDB();
  e.target.reset();
  renderKelolaAlat();
  showToast({ kind: 'success', message: `Alat baru ${name} (${code}) berhasil ditambahkan!` });
}

function toggleAktifAlat(id) {
  const tool = db.tools.find(t => t.id === id);
  if (!tool) return;
  tool.active = tool.active ? 0 : 1;
  saveDB();
  renderKelolaAlat();
  showToast({ kind: 'warn', message: `Status alat ${tool.name} berhasil diperbarui.` });
}

function renderKelolaTeknisi() {
  const grid = document.getElementById('kelola-teknisi-grid');
  grid.innerHTML = db.technicians.map(tech => `
    <div class="p-4 rounded-2xl border-2 border-slate-200 flex items-center justify-between ${!tech.active ? 'opacity-40 bg-slate-100' : 'bg-white'}">
      <div>
        <div class="font-black text-xl text-slate-900">${tech.name}</div>
        <div class="text-sm font-bold text-slate-500">${tech.section}</div>
      </div>
      <button onclick="toggleAktifTeknisi(${tech.id})" class="text-xs font-bold px-3 py-1.5 rounded-xl border border-slate-300 hover:bg-slate-100 transition-colors cursor-pointer">
        ${tech.active ? 'Nonaktifkan' : 'Aktifkan'}
      </button>
    </div>
  `).join('');
}

function tambahTeknisi(e) {
  e.preventDefault();
  const name = document.getElementById('input-tech-name').value.trim();
  const section = document.getElementById('input-tech-sec').value.trim();

  db.technicians.push({
    id: Date.now(),
    name,
    section,
    active: 1
  });

  saveDB();
  e.target.reset();
  renderKelolaTeknisi();
  showToast({ kind: 'success', message: `Teknisi ${name} (${section}) berhasil ditambahkan!` });
}

function toggleAktifTeknisi(id) {
  const tech = db.technicians.find(t => t.id === id);
  if (!tech) return;
  tech.active = tech.active ? 0 : 1;
  saveDB();
  renderKelolaTeknisi();
  showToast({ kind: 'warn', message: `Status teknisi ${tech.name} berhasil diubah.` });
}

function simpanPengaturan() {
  const hours = parseInt(document.getElementById('setting-max-hours').value) || 24;
  const pagi = document.getElementById('setting-shift-pagi').value || '07:00';
  const malam = document.getElementById('setting-shift-malam').value || '19:00';

  db.settings.default_max_hours = hours;
  db.settings.shift_pagi_start = pagi;
  db.settings.shift_malam_start = malam;
  saveDB();

  showToast({ kind: 'success', message: 'Pengaturan sistem berhasil disimpan.' });
}

// Clock updater
function startClock() {
  const el = document.getElementById('live-clock');
  const badge = document.getElementById('shift-badge');
  function tick() {
    const now = new Date();
    const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
    const dayName = days[now.getDay()];
    const day = now.getDate();
    const month = months[now.getMonth()];
    const h = String(now.getHours()).padStart(2, '0');
    const m = String(now.getMinutes()).padStart(2, '0');
    if (el) el.textContent = `${dayName}, ${day} ${month} - ${h}:${m} WIB`;
    const shift = getShift(now);
    if (badge) badge.textContent = `Shift ${shift === 'pagi' ? 'Pagi' : 'Malam'}`;
  }
  tick();
  setInterval(tick, 10000);
}

// Initialize on Load
document.addEventListener('DOMContentLoaded', () => {
  startClock();
  renderBeranda();
});

