package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func stringOfIntsToArrayOfInts(a string) []int32 {
	b := strings.Split(a, " ")
	N := len(b)
	B := make([]int32, N)
	var x int
	var s string

	for n := 0; n < N; n++ {
		s = b[n]
		x, _ = strconv.Atoi(s)
		B[n] = int32(x)
	}

	return B
}

// Complete the cost function below.
func cost(B []int32) int32 {
	N := len(B)

	if N <= 1 {
		return int32(0)
	} else if N == 2 {
		return int32(math.Max(float64(B[0]-int32(1)), float64(B[1]-int32(1))))
	}

	S := make([]int32, 3)
	S[0] = 0
	S[1] = B[N-1] - int32(1)

	// This won't run if N == 3, which is fine
	for n := 2; n < N-1; n++ {
		S[n%3] = int32(math.Max(float64(S[(n-1)%3]), float64(2)*(float64(B[N-n])-float64(1))+float64(S[(n-2)%3])))
	}

	return int32(math.Max(float64(B[0])-float64(1)+float64(S[(N-2)%3]), float64(2)*(float64(B[1])-float64(1))+float64(S[(N-3)%3])))
}

func main() {

	var B []int32
	var a string
	in := bufio.NewReader(os.Stdin)

	fmt.Println("Given an array B of positive integers, a 'tower array' of B is a new array A (of positive integers) with the same size as B such that, for all i, 1 <= A[i] <= B[i].")
	fmt.Println()
	fmt.Println("For any tower array A of the array B, define S(A) to be the sum of the absolute differences of consecutive elements of A, that is,")
	fmt.Println("S(A) = |A[1] - A[0]| + |A[2] - A[1]| + ... + |A[n-1] - A[n-2]|, where n is the size of B (and A).")
	fmt.Println()
	fmt.Println("The 'cost' of the array B is the minimum of S(A) taken over all tower arrays A of B.")
	fmt.Println()
	fmt.Println("Using dynamic programming, this program calculates the cost of user inputted arrays of positive integers in linear time and constant space complexity.")
	fmt.Println()
	fmt.Println()
	fmt.Println()

	for true {
		fmt.Println("Enter a list of single-space-separated positive integers, then press enter. To exit, enter -1.")
		a, _ = in.ReadString('\n')
		a = a[:len(a)-2]
		a = strings.TrimSpace(a)
		if strings.Compare(a, "-1") == 0 {
			break
		}

		B = stringOfIntsToArrayOfInts(a)
		fmt.Printf("Cost of this array: %v", cost(B))
		fmt.Println()
		fmt.Println()
	}

	fmt.Println("Thanks for using my program! =)")
}
