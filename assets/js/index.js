import { CSSGlobalVariables } from './css-global/css-global-variables.js'
let cssVar = new CSSGlobalVariables();

export function lineHeightAdust(step) {
    let old = parseFloat(cssVar.defspace);
    let sic = parseFloat(step);
    let nv = old + sic;
    let final = nv.toString();
    cssVar.defspace = final;
    console.log(cssVar.defspace)
}

export function fontSizeAdjust(step) {
    let current = parseInt(cssVar.fontsize.slice(0, 2));
    let ourStep = parseInt(step);
    let final = current + ourStep;
    if (final > 72) {
        final = 72;
    }
    cssVar.fontsize = final.toString() + "pt";
    console.log(cssVar.fontsize.slice(0, 2));
}


window.lineHeightAdust = lineHeightAdust;
window.fontSizeAdjust = fontSizeAdjust;