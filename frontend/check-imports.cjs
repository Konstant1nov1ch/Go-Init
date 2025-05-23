const fs = require('fs');
const path = require('path');

// Читаем файл
const filePath = path.join(__dirname, 'src/pages/CreateTemplate.tsx');
const content = fs.readFileSync(filePath, 'utf8');

// Ищем импорты
const importRegex = /import\s+{([^}]+)}\s+from\s+['"]([^'"]+)['"]/g;
let match;

console.log(`Checking imports in ${filePath}:\n`);

while ((match = importRegex.exec(content)) !== null) {
  const imports = match[1].trim().split(',').map(s => s.trim());
  const source = match[2];
  
  console.log(`From "${source}":`);
  imports.forEach(imp => console.log(`  - ${imp}`));
  console.log('');
}

// Специально ищем GET_TEMPLATE_STATUS
if (content.includes('GET_TEMPLATE_STATUS')) {
  console.log('!!! WARNING: GET_TEMPLATE_STATUS is found in the file');
  
  // Находим строку с GET_TEMPLATE_STATUS
  const lines = content.split('\n');
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].includes('GET_TEMPLATE_STATUS')) {
      console.log(`Line ${i+1}: ${lines[i]}`);
    }
  }
} 