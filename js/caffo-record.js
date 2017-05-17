;(function(){
    var Z = {
        createdCallback: function() {
            if(this.initialized) return;
            this.appendChild(document.importNode(this.template.content, true));

            if(!this.hasAttribute("recommended")) {
                let recommendedElement = this.querySelector('.recommended');
                recommendedElement.parentElement.removeChild(recommendedElement);
            }
            
            let imageElement = this.querySelector("image");
            let imageURL = this.getAttribute("image");
            imageElement.setAttributeNS('http://www.w3.org/1999/xlink', 'href', imageURL);

            ["artist", "title", "year", "genre"].forEach((name) => {
                let value = this.getAttribute(name);
                switch(name) {
                    case "title": value = `“${value}”`; this.removeAttribute("title"); break;
                    case "genre": value = "#"+value; break;
                }
                this.querySelector("."+name).textContent = value;
            });
            
            [].forEach.call(this.querySelectorAll('fit-text'), (fitText) => {
                fitText.connectedCallback();
                fitText.resize();
            });

            this.resize = () => {
                [].forEach.call(this.querySelectorAll('fit-text'), (fitText) => fitText.resize());
            };
            this.setAttribute("resolved","");
            this.initialized = true;
        },
    };

    Register("CaffoRecord", "caffo-record", Z, `
    <div class='record'>
      <div class='recommended'><img src='/svg/star-fill--white.svg'>TRY ME</div>
      <div class='record-vinyl'></div>
      <svg class='record-cover' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'>
        <image width='100%' height='100%' />
      </svg>
      <div class='record-top-label'>
        <fit-text class='artist'></fit-text>
        <fit-text class='title'></fit-text>
      </div>
      <div class='record-center-label'></div>
      <div class='record-bottom-label'>
        <span class='year'></span>
        <span class='genre'></span>
      </div>
    </div>`
    );
})();
