/*
The reducer has the purpose to filter out all RPM packages of a repo which have 0 chance to be involved in a solution.
This is a fast preflight-filter which drastically reduces the amount of variables in the CNF equation for the sat solver.
*/
package reducer
