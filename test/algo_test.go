//author: richard
package test

import (
	"fmt"
	"testing"
)

//		   { 0 	(n=0)
//Fib(n) = { 1  (n=1)
//		   { Fib(n-2) + Fib(n-1) (n>=2)
func TestFib(t *testing.T) {
   grid := [][]byte {
   	[]byte{1, 1, 1, 1, 0},
	[]byte{1, 1, 0, 1, 0},
    []byte{1, 1, 0, 0, 0},
    []byte{0, 0, 0, 0, 0},
   }
   nums := numIslands(grid)
   t.Log(nums)
}

func numIslands(grid [][]byte) int {
	row := len(grid)
	column := len(grid[0])
	isLands := 0
	for i := 0; i < row; i++ {
		for j := 0; j < column; j++ {
			if grid[i][j] == 1 {
				isLands++
				dfs(grid, i, j)
			} else {
				continue
			}
		}
	}
	return isLands
}

func dfs(grid [][]byte, r int, c int) {
	row := len(grid)
	column := len(grid[0])
	if r < 0 || c < 0 || r > row || c > column || grid[r][c] == 0 {
		return
	}
	//标记
	grid[r][c] = 0
	dfs(grid, r - 1, c)
	dfs(grid, r + 1, c)
	dfs(grid, r, c - 1)
	dfs(grid, r, c + 1)
}

func TestStdin(t *testing.T) {
	var first, last string
	_, err := fmt.Scanln(&first, &last)
	if err != nil {
		t.Error(err.Error())
	}
}