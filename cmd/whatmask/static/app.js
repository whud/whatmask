const input = document.getElementById('input');
const errorDiv = document.getElementById('error');
const resultsDiv = document.getElementById('results');
let timer = null;

input.addEventListener('input', function () {
  clearTimeout(timer);
  const val = this.value.trim();
  if (!val) {
    errorDiv.textContent = '';
    resultsDiv.innerHTML = '';
    return;
  }
  timer = setTimeout(() => fetchCalc(val), 300);
});

async function fetchCalc(val) {
  errorDiv.textContent = '';
  resultsDiv.innerHTML = '';
  try {
    const res = await fetch('/api/calc?input=' + encodeURIComponent(val));
    const data = await res.json();
    if (data.error) {
      errorDiv.textContent = data.error;
      return;
    }
    renderResults(data);
  } catch {
    errorDiv.textContent = 'request failed';
  }
}

function renderResults(data) {
  const rows = [];

  if (data.mode === 'network6') {
    rows.push(
      ['Address (compressed)', data.address],
      ['Address (expanded)', data.address_full, true],
      ['Prefix Length', '/' + data.cidr],
      ['Network (compressed)', data.network],
      ['Network (expanded)', data.network_full, true],
      ['Last (compressed)', data.last],
      ['Last (expanded)', data.last_full, true],
      ['Total Addresses', data.total],
      ['Type', data.type]
    );
  } else {
    if (data.mode === 'network') {
      rows.push(['IP Entered', data.address]);
    }
    rows.push(
      ['CIDR', '/' + data.cidr],
      ['Netmask', data.netmask],
      ['Netmask (hex)', data.hex],
      ['Wildcard Bits', data.wildcard]
    );
    if (data.mode === 'network') {
      rows.push(
        ['Network Address', data.network],
        ['Broadcast Address', data.broadcast],
        ['First Usable IP', data.first],
        ['Last Usable IP', data.last]
      );
    }
    rows.push(['Usable IPs', data.usable.toString()]);
  }

  const table = document.createElement('table');
  for (const [label, value, muted] of rows) {
    const tr = document.createElement('tr');
    if (muted) tr.classList.add('muted');
    const th = document.createElement('th');
    th.textContent = label;
    const td = document.createElement('td');
    td.textContent = value;
    tr.appendChild(th);
    tr.appendChild(td);
    table.appendChild(tr);
  }
  resultsDiv.appendChild(table);
}

input.focus();
