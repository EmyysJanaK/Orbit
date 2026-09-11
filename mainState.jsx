import { useState } from "react";

function Counter() {
    const[count, setCount] = useState(0);

    return(
        <div>
            <p>Current count:{count}</p>
            <button onClick={()=>setCount(count+1)}> 
                Increase
            </button>
        </div>
    );
    //This is a JavaScript arrow function "()=>" (or anonymous callback function). 
    // It acts as a wrapper that prevents the code inside from running 
    // immediately when the component renders. 
    // The code only executes after the user actually clicks.

    // This is React's event listener. ""onCick={...}""
    // It listens for a click event on the element and triggers the function provided inside the curly braces.

}