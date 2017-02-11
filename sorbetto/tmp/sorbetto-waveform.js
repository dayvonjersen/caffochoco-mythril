;(function(){

function createColorWaveform(srcImageElement, destImageElement, hexcolor) {
    var canvasElement = document.createElement("canvas");
    canvasElement.width = srcImageElement.width;
    canvasElement.height = srcImageElement.height;
    var canvasCtx = canvasElement.getContext('2d');
    canvasCtx.drawImage(srcImageElement, 0, 0);
    var imageData = canvasCtx.getImageData(0, 0, canvasElement.width, canvasElement.height);
    var data = imageData.data;
    var c = parseInt(hexcolor.substr(1), 16);
    var r = c>>16&0xff,
    g = c>>8&0xff,
    b = c&0xff;
    for(var i = 0; i < data.length; i+= 4) {
        if(data[i+3] == 255) {
            data[i] = r;
            data[i+1] = g;
            data[i+2] = b;
        }
    }
    canvasCtx.putImageData(imageData, 0, 0);
    destImageElement.src = canvasElement.toDataURL("image/png");
}
const minDistance = 0;
const maxAngle = Infinity;
const DEGREES_IN_RADIAN = 180 / Math.PI;
var Z = {
    createdCallback: function() {
    },
    attachedCallback: function() {
        var imageElement = this.querySelector('img');
        this.innerHTML="";
        if(!imageElement) return;

        var tpl = document.importNode(this.template.content, true);
        var containerElement = tpl.querySelector('.container');
        var topImageElement = tpl.querySelector('.image .top'); 
        var bottomImageElement = tpl.querySelector('.image .bottom'); 
        
        var topColor = this.getAttribute('data-top-color') || "#acaaaa";
        var bottomColor = this.getAttribute('data-bottom-color') || "#6d6a6a";
        var waveformImage = new Image();
        waveformImage.crossOrigin = 'anonymous';
        waveformImage.src = imageElement.src;
        waveformImage.onload = () => {
        createColorWaveform(waveformImage, topImageElement, topColor);
        createColorWaveform(waveformImage, bottomImageElement, bottomColor);
        };

        var state = {
            x: NaN,
            y: NaN,
            isDragging: false,
            position: 0
        };

        function updatePosition(e) {
            if(containerElement.disabled) return;

            [].forEach.call(containerElement.querySelectorAll("[data-style-property]"), function(childElement) {
                var pos = (childElement.dataset.invertPosition ? state.position - 1 : state.position).toPrecision(14) * 100;
                childElement.style[childElement.dataset.styleProperty] = pos+"%";
            });

            if("undefined" !== typeof e)  this.dispatchEvent(new Event("update"));
        };

        function takeoffeveryZig(e) {
            if(e) e.preventDefault();
            const isTouch = 'touches' in e;

            let pageX, pageY;
            if(isTouch) {
                pageX = e.touches[0].pageX;
                pageY = e.touches[0].pageY;
            } else {
                pageX = e.pageX;
                pageY = e.pageY;
            }

            state.x = pageX;
            state.y = pageY;

            document.addEventListener('mousemove', moveZig);
            document.addEventListener('touchmove', moveZig);
            document.addEventListener('mouseup', forgreatJustice);
            document.addEventListener('touchend', forgreatJustice);
        };

        function moveZig(e) {
            const isTouch = 'touches' in e;

            let pageX, pageY;
            if(isTouch) {
                pageX = e.touches[0].pageX;
                pageY = e.touches[0].pageY;
            } else {
                pageX = e.pageX;
                pageY = e.pageY;
            }

            if(!state.isDragging && isTouch) {
                const dx = state.x - pageX;
                const dy = state.y - pageY;

                const angle = Math.atan(dy/dx) * DEGREES_IN_RADIAN;
                const distance = Math.sqrt(dx * dx + dy * dy);

                let isDragging = distance >= minDistance;
                if(isDragging && Math.abs(angle) > maxAngle) {
                    // They're trying to scroll vertically
                    forgreatJustice();
                    return;
                } else if(!isDragging) {
                    return;
                }
                state.isDragging = isDragging;
            }

            const rect = containerElement.getBoundingClientRect();
            state.position = Math.max(Math.min( ((pageX - rect.left) / rect.width) , 1), 0);
            updatePosition(e);
        };

        function forgreatJustice() {    
            document.removeEventListener('mousemove', moveZig);
            document.removeEventListener('touchmove', moveZig);
            document.removeEventListener('mouseup', forgreatJustice);
            document.removeEventListener('touchend', forgreatJustice);

            state.isDragging = false;
            state.x = NaN;
            state.y = NaN;
            updatePosition();
        };

        function youknowwhatyoudoing(e) {
            takeoffeveryZig(e);
            moveZig(e);
        };

        containerElement.ondragstart = takeoffeveryZig;
        containerElement.ontouchstart = youknowwhatyoudoing;
        containerElement.onmousedown = youknowwhatyoudoing;
        containerElement.ondrag = moveZig;
        containerElement.ondragend = forgreatJustice;
        containerElement.ontouchend = forgreatJustice;
        containerElement.onmouseup = forgreatJustice;

        // set position to 0 on init
        updatePosition();

        // provide hooks to get/set position externally
        this.getPosition = () => state.position;
        this.setPosition = (p) => {
            state.position = p;
            updatePosition();
        };

        Object.defineProperty(this, "value", {
            get: function() {
                return this.getPosition();
            },
            set: function(p) {
                return this.setPosition(p);
            },
        });

        this.appendChild(tpl);
    },
    childListChangedCallback: function(removedNodes, addedNodes) {
        console.table(addedNodes);
        this.attachedCallback();
    }    
};   
Register("SorbettoWaveform", "sorbetto-waveform", Z, " <div class=\"container\"> <div data-style-property=\"left\" class=\"slider_wrapper\"> <div class=\"slider\"></div> </div> <!-- \"top\" image --> <div data-style-property=\"left\" data-invert-position=\"1\" class=\"image_wrapper\" style=\"margin-right:-2px\"> <div data-style-property=\"right\" data-invert-position=\"1\" class=\"image\"> <img class=\"top\"></img> </div> </div> <!-- \"bottom\" image --> <div data-style-property=\"left\" data-invert-position=\"1\" class=\"image_wrapper\" style=\"margin-left:-2px\"> <div data-style-property=\"right\" class=\"image\"> <img class=\"bottom\"></img> </div> </div> </div> ");
})();