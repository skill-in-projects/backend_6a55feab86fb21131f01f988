package main

import "testing"

func add(a, b int) int {
    return a + b
}

func TestAdd_ReturnsSumOfTwoNumbers(t *testing.T) {
    if got := add(2, 3); got != 5 {
        t.Errorf("add(2, 3) = %d; want 5", got)
    }
}

func TestAdd_WithNegativeNumbers_ReturnsCorrectSum(t *testing.T) {
    if got := add(2, -3); got != -1 {
        t.Errorf("add(2, -3) = %d; want -1", got)
    }
}
