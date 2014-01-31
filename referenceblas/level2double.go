package referenceblas

import "github.com/gonum/blas"

// See http://www.netlib.org/lapack/explore-html/d4/de1/_l_i_c_e_n_s_e_source.html
// for more license information

//var _ blas.Float64Level2 = Blasser

// TODO: Need to think about loops when doing row-major. Change after tests?

const (
	badOrder     string = "referenceblas: illegal order"
	mLT0         string = "referenceblas: m < 0"
	nLT0         string = "referenceblas: m < 0"
	badUplo      string = "referenceblas: illegal triangularization"
	badTranspose string = "referenceblas: illegal transpose"
	badDiag      string = "referenceblas: illegal diag"
	badLdaRow    string = "lda must be greater than max(1,n) for row major"
	badLdaCol    string = "lda must be greater than max(1,m) for col major"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Dgemv computes y = alpha*a*x + beta*y if tA = blas.NoTrans
// or alpha*A^T*x + beta*y if tA = blas.Trans or blas.ConjTrans
func (b Blas) Dgemv(o blas.Order, tA blas.Transpose, m, n int, alpha float64, a []float64, lda int, x []float64, incX int, beta float64, y []float64, incY int) {
	if o != blas.RowMajor && o != blas.ColMajor {
		panic(badOrder)
	}
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if m < 0 {
		panic(mLT0)
	}
	if n < 0 {
		panic(nLT0)
	}
	if o == blas.RowMajor {
		if lda < max(1, n) {
			panic(badLdaRow)
		}
	} else {
		if lda < max(1, m) {
			panic(badLdaCol)
		}
	}
	if incX == 0 {
		panic(zeroInc)
	}
	if incY == 0 {
		panic(zeroInc)
	}

	// Quick return if possible
	if m == 0 || n == 0 || (alpha == 0 && beta == 1) {
		return
	}

	// Set up indexes
	lenX := m
	lenY := n
	if tA == blas.NoTrans {
		lenX = n
		lenY = m
	}
	var kx, ky int
	if incX > 0 {
		kx = 0
	} else {
		kx = -(lenX - 1) * incX
	}
	if incY > 0 {
		ky = 0
	} else {
		ky = -(lenY - 1) * incY
	}

	// First form y := beta * y
	if incY > 0 {
		b.Dscal(lenY, beta, y, incY)
	} else {
		b.Dscal(lenY, beta, y, -incY)
	}

	if alpha == 0 {
		return
	}

	// Form y := alpha * A * x + y
	switch {

	default:
		panic("shouldn't be here")

	case o == blas.RowMajor && tA == blas.NoTrans:
		iy := ky
		for i := 0; i < m; i++ {
			jx := kx
			var temp float64
			for j := 0; j < n; j++ {
				temp += a[lda*i+j] * x[jx]
				jx += incX
			}
			y[iy] += alpha * temp
			iy += incY
		}
	case o == blas.RowMajor && (tA == blas.Trans || tA == blas.ConjTrans):
		ix := kx
		for i := 0; i < m; i++ {
			jy := ky
			tmp := alpha * x[ix]
			for j := 0; j < n; j++ {
				y[jy] += a[lda*i+j] * tmp
				jy += incY
			}
			ix += incX
		}

	case o == blas.ColMajor && tA == blas.NoTrans:
		jx := kx
		for j := 0; j < n; j++ {
			temp := alpha * x[jx]
			iy := ky
			for i := 0; i < m; i++ {
				y[iy] += temp * a[lda*j+i]
				iy += incY
			}
			jx += incX
		}
	case o == blas.ColMajor && (tA == blas.Trans || tA == blas.ConjTrans):
		jy := ky
		for j := 0; j < n; j++ {
			var temp float64
			ix := kx
			for i := 0; i < m; i++ {
				temp += a[lda*j+i] * x[ix]
				ix += incX
			}
			y[jy] += alpha * temp
			jy += incY
		}
	}
}

// Dger   performs the rank 1 operation
//    A := alpha*x*y**T + A,
// where alpha is a scalar, x is an m element vector, y is an n element
// vector and A is an m by n matrix.
func (Blas) Dger(o blas.Order, m, n int, alpha float64, x []float64, incX int, y []float64, incY int, a []float64, lda int) {
	// Check inputs
	if o != blas.RowMajor && o != blas.ColMajor {
		panic(badOrder)
	}
	if m < 0 {
		panic("m < 0")
	}
	if n < 0 {
		panic(negativeN)
	}
	if incX == 0 {
		panic(zeroInc)
	}
	if incY == 0 {
		panic(zeroInc)
	}
	if o == blas.RowMajor {
		if lda < max(1, n) {
			panic(badLdaRow)
		}
	} else {
		if lda < max(1, m) {
			panic(badLdaCol)
		}
	}
	// Quick return if possible
	if m == 0 || n == 0 || alpha == 0 {
		return
	}

	var ky, kx int
	if incY > 0 {
		ky = 0
	} else {
		ky = -(n - 1) * incY
	}

	if incY > 0 {
		kx = 0
	} else {
		kx = -(m - 1) * incX
	}

	switch o {
	default:
		panic("should not be here")
	case blas.RowMajor:
		ix := kx
		for i := 0; i < m; i++ {
			if x[ix] == 0 {
				ix += incX
				continue
			}
			tmp := alpha * x[ix]
			jy := ky
			for j := 0; j < n; j++ {
				a[i*lda+j] += y[jy] * tmp
				jy += incY
			}
			ix += incX
		}
	case blas.ColMajor:
		jy := ky
		for j := 0; j < n; j++ {
			if y[jy] == 0 {
				jy += incY
				continue
			}
			tmp := alpha * y[jy]
			ix := kx
			for i := 0; i < m; i++ {
				a[j*lda+i] += x[ix] * tmp
				ix += incX
			}
			jy += incY
		}

	}
}

func (b Blas) Dgbmv(o blas.Order, tA blas.Transpose, m, n, kL, kU int, alpha float64, a []float64, lda int, x []float64, incX int, beta float64, y []float64, incY int) {
	if o != blas.RowMajor && o != blas.ColMajor {
		panic(badOrder)
	}
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if m < 0 {
		panic(mLT0)
	}
	if n < 0 {
		panic(nLT0)
	}
	if o == blas.RowMajor {
		if lda < max(1, n) {
			panic(badLdaRow)
		}
	} else {
		if lda < max(1, m) {
			panic(badLdaCol)
		}
	}
	if incX == 0 {
		panic(zeroInc)
	}
	if incY == 0 {
		panic(zeroInc)
	}

	// Quick return if possible
	if m == 0 || n == 0 || (alpha == 0 && beta == 1) {
		return
	}

	// Set up indexes
	lenX := m
	lenY := n
	if tA == blas.NoTrans {
		lenX = n
		lenY = m
	}
	var kx, ky int
	if incX > 0 {
		kx = 0
	} else {
		kx = -(lenX - 1) * incX
	}
	if incY > 0 {
		ky = 0
	} else {
		ky = -(lenY - 1) * incY
	}

	// First form y := beta * y
	if incY > 0 {
		b.Dscal(lenY, beta, y, incY)
	} else {
		b.Dscal(lenY, beta, y, -incY)
	}

	if alpha == 0 {
		return
	}
	kup1 := kU + 1

	if o == blas.RowMajor {
		m, n = n, m
		if tA == blas.NoTrans {
			tA = blas.Trans
		} else {
			tA = blas.NoTrans
		}
	}

	if tA == blas.NoTrans {
		jx := kx
		if incY == 1 {
			for j := 0; j < n; j++ {
				if x[jx] != 0 {
					temp := alpha * x[jx]
					k := kup1 - j
					for i := max(0, j-kU-1); i < min(m, j+kL); i++ {
						y[i] += temp * a[k+i+j*lda]
					}
				}
				jx += incX
			}
		} else {
			for j := 0; j < n; j++ {
				if x[jx] != 0 {
					temp := alpha * x[jx]
					iy := ky
					k := kup1 - j
					for i := max(0, j-kU-1); i < min(m, j+kL); i++ {
						y[iy] += temp * a[k+i+j*lda]
						iy += incY
					}
				}
				jx += incX
				if j > kU {
					ky += incY
				}
			}
		}
	} else {
		jy := ky
		if incX == 1 {
			for j := 0; j < n; j++ {
				temp := 0.0
				k := kup1 - j
				for i := max(0, j-kU-1); i < min(m, j+kL); i++ {
					temp += a[k+i+j*lda] * x[i]
				}
				y[jy] += alpha * temp
				jy += incX
			}
		} else {
			for j := 0; j < n; j++ {
				temp := 0.0
				ix := kx
				k := kup1 - j
				for i := max(0, j-kU-1); i < min(m, j+kL); i++ {
					temp += a[k+i+j*lda] * x[ix]
					ix += incX
				}
				y[jy] += alpha * temp
				jy += incX
				if j > kU {
					kx += incX
				}
			}
		}
	}
}

/*
// DTRMV  performs one of the matrix-vector operations
// 		x := A*x,   or   x := A**T*x,
// where x is an n element vector and  A is an n by n unit, or non-unit,
// upper or lower triangular matrix.
func (Blas) Dtrmv(o blas.Order, ul blas.Uplo, tA blas.Transpose, d blas.Diag, n int, a []float64, lda int, x []float64, incX int) {
	// Verify inputs
	if o != blas.RowMajor && o != blas.ColMajor {
		panic(badOrder)
	}
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if d != blas.NonUnit && d != blas.Unit {
		panic(badDiag)
	}
	if n < 0 {
		panic(nLT0)
	}
	if lda > n && lda > 1 {
		panic("blas: lda must be less than max(1,n)")
	}
	if incX == 0 {
		panic(zeroInc)
	}
	if n == 0 {
		return
	}
	var kx int
	if incX <= 0 {
		kx = -(n - 1) * incX
	}
	switch {
	default:
		panic("not yet implemented")
	case o == blas.RowMajor && tA == blas.NoTrans && ul == blas.Upper:
		jx := kx
		for j := 0; j < n; j++ {
			if x[jx] != 0 {
				temp := x[jx]
				ix := kx
				for i := 0; i < j-1; i++ {
					x[ix] += temp * a[lda*i+j]
					ix += incX
				}
				if d == blas.NonUnit {
					x[jx] *= a[lda*j+j]
				}
			}
			jx += incX
		}
	case o == blas.RowMajor && tA == blas.NoTrans && ul == blas.Lower:
		kx += (n - 1) * incX
		jx := kx
		for j := n - 1; j >= 0; j-- {
			if x[jx] != 0 {
				tmp := x[jx]
				ix := kx
				for i := n - 1; i >= j; i-- {
					x[ix] += tmp * a[lda*i+j]
				}
				if d == blas.NonUnit {
					x[jx] *= a[lda*j+j]
				}
			}
		}
	case o == blas.RowMajor && (tA == blas.Trans || tA == blas.ConjTrans) && ul == blas.Upper:
		jx := kx + (n-1)*incX
		for j := n - 1; j >= 0; j-- {
			tmp := x[jx]
			ix := jx
			if d == blas.NonUnit {
				tmp *= a[lda*j+j]
			}
			for i := j - 2; j >= 0; j-- {
				ix -= incX
				tmp += a[lda*i+j] * x[ix]
			}
			x[jx] = tmp
			jx -= incX
		}
	case o == blas.RowMajor && (tA == blas.Trans || tA == blas.ConjTrans) && ul == blas.Lower:
		jx := kx
		for j := 0; j < n; j++ {
			tmp := x[jx]
			ix := jx
			if d == blas.NonUnit {
				tmp *= a[lda*j+j]
				for i := j; i < n; i++ {
					ix += incX
					tmp += a[lda*i+j] * x[ix]
				}
				x[jx] = tmp
				jx += incX
			}
		}
	}
}

// Dtrsv  solves one of the systems of equations
//    A*x = b,   or   A**T*x = b,
// where b and x are n element vectors and A is an n by n unit, or
// non-unit, upper or lower triangular matrix.
//
// No test for singularity or near-singularity is included in this
// routine. Such tests must be performed before calling this routine.
func (Blas) Dtrsv(o blas.Order, ul blas.Uplo, tA blas.Transpose, d blas.Diag, n int, a []float64, lda int, x []float64, incX int) {
	// Test the input parameters
	// Verify inputs
	if o != blas.RowMajor && o != blas.ColMajor {
		panic(badOrder)
	}
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if d != blas.NonUnit && d != blas.Unit {
		panic(badDiag)
	}
	if n < 0 {
		panic(nLT0)
	}
	if lda > n && lda > 1 {
		panic("blas: lda must be less than max(1,n)")
	}
	if incX == 0 {
		panic(zeroInc)
	}
	// Quick return if possible
	if n == 0 {
		return
	}

	var kx int
	if incX < 0 {
		kx = -(n - 1) * incX
	}

	switch {
	default:
		panic("col major not yet coded")
	case o == blas.RowMajor && tA == blas.NoTrans && ul == blas.Upper:
		jx := kx + (n-1)*incX
		for j := n; j >= 0; j-- {
			if x[jx] != 0 {
				if d == blas.NonUnit {
					x[jx] /= a[lda*j+j]
				}
				tmp := x[jx]
				ix := jx
				for i := j - 2; i >= 0; i-- {
					ix -= incX
					x[ix] -= tmp * a[lda*i+j]
				}
			}
			jx -= incX
		}
	case o == blas.RowMajor && tA == blas.NoTrans && ul == blas.Lower:
		jx := kx
		for j := 0; j < n; j++ {
			if x[jx] != 0 {
				if d == blas.NonUnit {
					x[jx] /= a[lda*j+j]
				}
				tmp := x[jx]
				ix := jx
				for i := j; i < n; j++ {
					ix += incX
					x[ix] -= tmp * a[lda*i+j]
				}
			}
			jx += incX
		}
	case o == blas.RowMajor && (tA == blas.Trans || tA == blas.ConjTrans) && ul == blas.Upper:
		jx := kx
		for j := 0; j < n; j++ {
			tmp := x[jx]
			ix := kx
			for i := 0; i < j-1; i++ {
				tmp -= a[lda*i+j] * x[ix]
				ix += incX
			}
			if d == blas.NonUnit {
				tmp /= a[lda*j+j]
			}
			x[jx] = tmp
			jx += incX
		}
	case o == blas.RowMajor && (tA == blas.Trans || tA == blas.ConjTrans) && ul == blas.Lower:
		kx += (n - 1) * incX
		jx := kx
		for j := n - 1; j >= 0; j-- {
			tmp := x[jx]
			ix := kx
			for i := n - 1; i >= j; i-- {
				tmp -= a[lda*i+j] * x[ix]
				ix -= incX
			}
			if d == blas.NonUnit {
				tmp /= a[lda*j+j]
				x[jx] = tmp
				jx -= incX
			}
		}
	}
}


// Dsymv  performs the matrix-vector  operation
//    y := alpha*A*x + beta*y,
// where alpha and beta are scalars, x and y are n element vectors and
// A is an n by n symmetric matrix.
func (b Blas) Dsymv(o blas.Order, ul blas.Uplo, n int, alpha float64, a []float64, lda int, x []float64, incX int, beta float64, y []float64, incY int) {
	// Check inputs
	if o != blas.RowMajor && o != blas.ColMajor {
		panic(badOrder)
	}
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if n < 0 {
		panic(negativeN)
	}
	if lda > 1 && lda > n {
		panic(badLda)
	}
	if incX == 0 {
		panic(zeroInc)
	}
	if incY == 0 {
		panic(zeroInc)
	}
	// Quick return if possible
	if n == 0 || (alpha == 0 && beta == 1) {
		return
	}

	// Set up start points
	var kx, ky int
	if incX > 0 {
		kx = 1
	} else {
		kx = -(n - 1) * incX
	}
	if incY > 0 {
		ky = 1
	} else {
		ky = -(n - 1) * incY
	}

	// Form y = beta * y
	if beta != 1 {
		b.Dscal(n, beta, y, incY)
	}

	if alpha == 0 {
		return
	}

	// TODO: Need to think about changing the major and minor
	// looping when row major (help with cache misses)

	// Form y = Ax + y
	switch {
	default:
		panic("not yet coded")
	case o == blas.RowMajor && ul == blas.Upper:
		jx := kx
		jy := ky
		for j := 0; j < n; j++ {
			tmp1 := alpha * x[jx]
			var tmp2 float64
			ix := kx
			iy := ky
			for i := 0; i < j-2; i++ {
				y[iy] += tmp1 * a[i*lda+j]
				tmp2 += a[i*lda+j] * x[ix]
				ix += incX
				iy += incY
			}
			y[jy] += tmp1*a[j*lda+j] + alpha*tmp2
			jx += incX
			jy += incY
		}
	case o == blas.RowMajor && ul == blas.Lower:
		jx := kx
		jy := ky
		for j := 0; j < n; j++ {
			tmp1 := alpha * x[jx]
			var tmp2 float64
			y[jy] += tmp1 * a[j*lda+j]
			ix := jx
			iy := jy
			for i := j; i < n; i++ {
				ix += incX
				iy += incY
				y[iy] += tmp1 * a[i*lda+j]
				tmp2 += a[i*lda+j] * x[ix]
			}
			y[jy] += alpha * tmp2
			jx += incX
			jy += incY
		}
	}
}
*/
/*
// Level 2 routines.
        Dgbmv(o Order, tA Transpose, m, n, kL, kU int, alpha float64, a []float64, lda int, x []float64, incX int, beta float64, y []float64, incY int)
        Dtbmv(o Order, ul Uplo, tA Transpose, d Diag, n, k int, a []float64, lda int, x []float64, incX int)
        Dtpmv(o Order, ul Uplo, tA Transpose, d Diag, n int, ap []float64, x []float64, incX int)
        Dtbsv(o Order, ul Uplo, tA Transpose, d Diag, n, k int, a []float64, lda int, x []float64, incX int)
        Dtpsv(o Order, ul Uplo, tA Transpose, d Diag, n int, ap []float64, x []float64, incX int)
        Dsbmv(o Order, ul Uplo, n, k int, alpha float64, a []float64, lda int, x []float64, incX int, beta float64, y []float64, incY int)
        Dspmv(o Order, ul Uplo, n int, alpha float64, ap []float64, x []float64, incX int, beta float64, y []float64, incY int)
		Dspr(o Order, ul Uplo, n int, alpha float64, x []float64, incX int, ap []float64)
		Dspr2(o Order, ul Uplo, n int, alpha float64, x []float64, incX int, y []float64, incY int, a []float64)
        Dsyr(o Order, ul Uplo, n int, alpha float64, x []float64, incX int, a []float64, lda int)
        Dsyr2(o Order, ul Uplo, n int, alpha float64, x []float64, incX int, y []float64, incY int, a []float64, lda int)
*/
