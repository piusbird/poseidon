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
window.lineHeightAdust = lineHeightAdust;