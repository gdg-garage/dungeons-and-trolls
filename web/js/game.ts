import "../css/game.scss";
import {renderMap} from "./rendering";
import {processUserInput} from "./controls";

document.addEventListener("keydown", processUserInput)

renderMap();