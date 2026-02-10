// Auto-initialize asciinema players on page load
document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('.fluffy-demo').forEach(function(el) {
        var cast = el.getAttribute('data-cast');
        var cols = parseInt(el.getAttribute('data-cols')) || 80;
        var rows = parseInt(el.getAttribute('data-rows')) || 24;
        var speed = parseFloat(el.getAttribute('data-speed')) || 1.5;
        var theme = el.getAttribute('data-theme') || 'monokai';

        AsciinemaPlayer.create(cast, el, {
            cols: cols,
            rows: rows,
            autoPlay: true,
            loop: true,
            speed: speed,
            theme: theme,
            fit: 'width',
            terminalFontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace"
        });
    });
});
