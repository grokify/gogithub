// SVG and PNG export utilities

export function downloadSVG(svgElement: SVGElement, filename: string): void {
  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(svgElement);
  const blob = new Blob([svgString], { type: 'image/svg+xml' });
  downloadBlob(blob, filename);
}

export function downloadPNG(svgElement: SVGElement, filename: string, scale: number = 2): void {
  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(svgElement);

  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');
  if (!ctx) {
    console.error('Could not get canvas context');
    return;
  }

  const img = new Image();
  const svgBlob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
  const url = URL.createObjectURL(svgBlob);

  img.onload = () => {
    canvas.width = img.width * scale;
    canvas.height = img.height * scale;

    // Fill with background color based on theme
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    ctx.fillStyle = isDark ? '#0d1117' : '#ffffff';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    ctx.scale(scale, scale);
    ctx.drawImage(img, 0, 0);

    canvas.toBlob((blob) => {
      if (blob) {
        downloadBlob(blob, filename);
      }
      URL.revokeObjectURL(url);
    }, 'image/png');
  };

  img.onerror = () => {
    console.error('Error loading SVG for PNG conversion');
    URL.revokeObjectURL(url);
  };

  img.src = url;
}

function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
