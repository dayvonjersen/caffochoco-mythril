;(function(){

/**
 * All credit to https://github.com/jxnblk/fitter-happier-text
 */
var Z = {
    createdCallback: function() {
        var textContent = this.textContent.trim();
        this.innerHTML  = "";
        this.setAttribute('aria-label', textContent);

        var svgElement  = document.createElementNS('http://www.w3.org/2000/svg', 'svg'),
            textElement = document.createElementNS('http://www.w3.org/2000/svg', 'text');

        textElement.textContent = textContent;
        textElement.setAttribute('x', '50%');
        textElement.setAttribute('y', '20');

        svgElement.appendChild(textElement);

        this.appendChild(svgElement);
        this.resize = () => {
            var svgElement = this.querySelector('svg'),
            textElement = this.querySelector('text');
            if(!textElement) return;
            var w = textElement.offsetWidth || textElement.getComputedTextLength(),
                h = textElement.offsetHeight || 24;
            svgElement.setAttribute('viewBox', '0 0 ' + w + ' ' + h);
        }
        var that = this;
        window.addEventListener('resize', function() {
            that.resize();
        });
        window.addEventListener('WebComponentsReady', function() {
            that.resize();
        });
    },
    attachedCallback: function() {
        this.resize();
    },
    disconnectedCallback: function() {
        this.innerHTML = this.textContent.trim();
    }
};
    
Register("FitText", "fit-text", Z, "");
})();
