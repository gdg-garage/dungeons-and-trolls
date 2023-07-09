import "../css/game.scss";
import {renderMap} from "./rendering";
import {processUserInput} from "./controls";

document.addEventListener("keydown", processUserInput)

function render() {
    // todo nasty
    document.getElementById("game").innerHTML = "";
    renderMap();
    setTimeout(() => {  render(); }, 1000);
}

render();