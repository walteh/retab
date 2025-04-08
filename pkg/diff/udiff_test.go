package diff

// import (
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestProcessDiff(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		diffStr  string
// 		expected *ProcessedDiff
// 		wantErr  bool
// 	}{
// 		{
// 			name:     "empty diff",
// 			diffStr:  "",
// 			expected: nil,
// 			wantErr:  true,
// 		},
// 		{
// 			name: "simple one-line change",
// 			diffStr: `--- file1.txt
// +++ file2.txt
// @@ -1,1 +1,1 @@
// -old line
// +new line`,
// 			expected: &ProcessedDiff{
// 				OrigFile: "file1.txt",
// 				NewFile:  "file2.txt",
// 				Hunks: []DiffHunk{
// 					{
// 						Header: "@@ -1,1 +1,1 @@",
// 						Lines: []DiffLine{
// 							{
// 								Content: "old line",
// 								Type:    DiffLineRemoved,
// 								Changes: []DiffChange{
// 									{Text: "old", Type: DiffChangeRemoved},
// 									{Text: " line", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "new line",
// 								Type:    DiffLineAdded,
// 								Changes: []DiffChange{
// 									{Text: "new", Type: DiffChangeAdded},
// 									{Text: " line", Type: DiffChangeUnchanged},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "with character-level changes",
// 			diffStr: `--- file1.txt
// +++ file2.txt
// @@ -1,1 +1,1 @@
// -Hello world
// +Hello universe`,
// 			expected: &ProcessedDiff{
// 				OrigFile: "file1.txt",
// 				NewFile:  "file2.txt",
// 				Hunks: []DiffHunk{
// 					{
// 						Header: "@@ -1,1 +1,1 @@",
// 						Lines: []DiffLine{
// 							{
// 								Content: "Hello world",
// 								Type:    DiffLineRemoved,
// 								Changes: []DiffChange{
// 									{Text: "Hello ", Type: DiffChangeUnchanged},
// 									{Text: "world", Type: DiffChangeRemoved},
// 								},
// 							},
// 							{
// 								Content: "Hello universe",
// 								Type:    DiffLineAdded,
// 								Changes: []DiffChange{
// 									{Text: "Hello ", Type: DiffChangeUnchanged},
// 									{Text: "universe", Type: DiffChangeAdded},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "context lines",
// 			diffStr: `--- file1.txt
// +++ file2.txt
// @@ -1,3 +1,3 @@
//  line 1
// -line 2
// +updated line 2
//  line 3`,
// 			expected: &ProcessedDiff{
// 				OrigFile: "file1.txt",
// 				NewFile:  "file2.txt",
// 				Hunks: []DiffHunk{
// 					{
// 						Header: "@@ -1,3 +1,3 @@",
// 						Lines: []DiffLine{
// 							{
// 								Content: "line 1",
// 								Type:    DiffLineContext,
// 								Changes: []DiffChange{
// 									{Text: "line 1", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "line 2",
// 								Type:    DiffLineRemoved,
// 								Changes: []DiffChange{
// 									{Text: "line 2", Type: DiffChangeRemoved},
// 								},
// 							},
// 							{
// 								Content: "updated line 2",
// 								Type:    DiffLineAdded,
// 								Changes: []DiffChange{
// 									{Text: "updated ", Type: DiffChangeAdded},
// 									{Text: "line 2", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "line 3",
// 								Type:    DiffLineContext,
// 								Changes: []DiffChange{
// 									{Text: "line 3", Type: DiffChangeUnchanged},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "multiple hunks",
// 			diffStr: `--- file1.txt
// +++ file2.txt
// @@ -1,3 +1,3 @@
//  line 1
// -line 2
// +updated line 2
//  line 3
// @@ -10,3 +10,3 @@
//  line 10
// -line 11
// +updated line 11
//  line 12`,
// 			expected: &ProcessedDiff{
// 				OrigFile: "file1.txt",
// 				NewFile:  "file2.txt",
// 				Hunks: []DiffHunk{
// 					{
// 						Header: "@@ -1,3 +1,3 @@",
// 						Lines: []DiffLine{
// 							{
// 								Content: "line 1",
// 								Type:    DiffLineContext,
// 								Changes: []DiffChange{
// 									{Text: "line 1", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "line 2",
// 								Type:    DiffLineRemoved,
// 								Changes: []DiffChange{
// 									{Text: "line 2", Type: DiffChangeRemoved},
// 								},
// 							},
// 							{
// 								Content: "updated line 2",
// 								Type:    DiffLineAdded,
// 								Changes: []DiffChange{
// 									{Text: "updated ", Type: DiffChangeAdded},
// 									{Text: "line 2", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "line 3",
// 								Type:    DiffLineContext,
// 								Changes: []DiffChange{
// 									{Text: "line 3", Type: DiffChangeUnchanged},
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Header: "@@ -10,3 +10,3 @@",
// 						Lines: []DiffLine{
// 							{
// 								Content: "line 10",
// 								Type:    DiffLineContext,
// 								Changes: []DiffChange{
// 									{Text: "line 10", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "line 11",
// 								Type:    DiffLineRemoved,
// 								Changes: []DiffChange{
// 									{Text: "line 11", Type: DiffChangeRemoved},
// 								},
// 							},
// 							{
// 								Content: "updated line 11",
// 								Type:    DiffLineAdded,
// 								Changes: []DiffChange{
// 									{Text: "updated ", Type: DiffChangeAdded},
// 									{Text: "line 11", Type: DiffChangeUnchanged},
// 								},
// 							},
// 							{
// 								Content: "line 12",
// 								Type:    DiffLineContext,
// 								Changes: []DiffChange{
// 									{Text: "line 12", Type: DiffChangeUnchanged},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "multiple changes in one hunk",
// 			diffStr: `--- file1.txt
// +++ file2.txt
// @@ -1,4 +1,4 @@
//  line 1
// -old line 2
// +new line 2
// -old line 3
// +new line 3`,
// 			expected: &ProcessedDiff{
// 				OrigFile: "file1.txt",
// 				NewFile:  "file2.txt",
// 				Hunks: []DiffHunk{
// 					{
// 						Header: "@@ -1,4 +1,4 @@",
// 						Lines: []DiffLine{
// 							{
// 								Content: "line 1",
// 								Type:    DiffLineContext,
// 							},
// 							{
// 								Content: "old line 2",
// 								Type:    DiffLineRemoved,
// 							},
// 							{
// 								Content: "new line 2",
// 								Type:    DiffLineAdded,
// 							},
// 							{
// 								Content: "old line 3",
// 								Type:    DiffLineRemoved,
// 							},
// 							{
// 								Content: "new line 3",
// 								Type:    DiffLineAdded,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ud, err := ParseUnifiedDiff(tt.diffStr)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				return
// 			}
// 			require.NoError(t, err)

// 			actual, err := ud.ProcessDiff()
// 			require.NoError(t, err)

// 			RequireKnownValueEqual(t, tt.expected, actual)

// 		})
// 	}
// }

// // Helper function to combine changes into a string
// func combineChanges(changes []DiffChange) string {
// 	var result strings.Builder
// 	for _, c := range changes {
// 		result.WriteString(c.Text)
// 	}
// 	return result.String()
// }

// func TestProcessLineChanges(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		oldLine  string
// 		newLine  string
// 		lineType DiffLineType
// 		expected DiffLine
// 	}{
// 		{
// 			name:     "added line with character changes",
// 			oldLine:  "first line",
// 			newLine:  "first modified line",
// 			lineType: DiffLineAdded,
// 			expected: DiffLine{
// 				Content: "first modified line",
// 				Type:    DiffLineAdded,
// 				Changes: []DiffChange{
// 					{Text: "first ", Type: DiffChangeUnchanged},
// 					{Text: "modified ", Type: DiffChangeAdded},
// 					{Text: "line", Type: DiffChangeUnchanged},
// 				},
// 			},
// 		},
// 		{
// 			name:     "removed line with character changes",
// 			oldLine:  "line with number 54",
// 			newLine:  "line with number 5",
// 			lineType: DiffLineRemoved,
// 			expected: DiffLine{
// 				Content: "line with number 54",
// 				Type:    DiffLineRemoved,
// 				Changes: []DiffChange{
// 					{Text: "line with number 5", Type: DiffChangeUnchanged},
// 					{Text: "4", Type: DiffChangeRemoved},
// 				},
// 			},
// 		},
// 		{
// 			name:     "completely different lines",
// 			oldLine:  "totally different",
// 			newLine:  "completely changed",
// 			lineType: DiffLineAdded,
// 			expected: DiffLine{
// 				Content: "completely changed",
// 				Type:    DiffLineAdded,
// 				Changes: []DiffChange{
// 					{Text: "completely changed", Type: DiffChangeAdded},
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := processLineChanges(tt.oldLine, tt.newLine, tt.lineType)

// 			require.Equal(t, tt.expected, result)

// 			// // Check the basic properties
// 			// assert.Equal(t, tt.expected.Content, result.Content)
// 			// assert.Equal(t, tt.expected.Type, result.Type)

// 			// // For character-level changes, compare the combined content
// 			// // since different diffing algorithms may split content differently
// 			// expectedContent := combineChanges(tt.expected.Changes)
// 			// actualContent := combineChanges(result.Changes)
// 			// assert.Equal(t, expectedContent, actualContent)

// 			// // Check that we have changes that follow the desired pattern
// 			// if len(tt.expected.Changes) > 0 {
// 			// 	hasExpectedChangeTypes := false

// 			// 	if tt.lineType == DiffLineAdded {
// 			// 		for _, change := range result.Changes {
// 			// 			if change.Type == DiffChangeAdded {
// 			// 				hasExpectedChangeTypes = true
// 			// 				break
// 			// 			}
// 			// 		}
// 			// 		require.True(t, hasExpectedChangeTypes, "changes missing expected added change type: %v", result.Changes)
// 			// 	} else if tt.lineType == DiffLineRemoved {
// 			// 		for _, change := range result.Changes {
// 			// 			if change.Type == DiffChangeRemoved {
// 			// 				hasExpectedChangeTypes = true
// 			// 				break
// 			// 			}
// 			// 		}
// 			// 		require.True(t, hasExpectedChangeTypes, "changes missing expected removed change type: %v", result.Changes)
// 			// 	}
// 			// }

// 		})
// 	}
// }

// func TestGroupRelatedChanges(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		lines    []string
// 		expected []lineGroup
// 	}{
// 		{
// 			name: "simple grouped changes",
// 			lines: []string{
// 				"context line 1",
// 				"-removed line",
// 				"+added line",
// 				"context line 2",
// 			},
// 			expected: []lineGroup{
// 				{
// 					contextLines: []string{"context line 1"},
// 					oldLines:     []string{"removed line"},
// 					newLines:     []string{"added line"},
// 				},
// 				{
// 					contextLines: []string{"context line 2"},
// 					oldLines:     []string{},
// 					newLines:     []string{},
// 				},
// 			},
// 		},
// 		{
// 			name: "multiple change groups",
// 			lines: []string{
// 				"context line 1",
// 				"-removed line 1",
// 				"+added line 1",
// 				"context line 2",
// 				"-removed line 2",
// 				"+added line 2",
// 				"context line 3",
// 			},
// 			expected: []lineGroup{
// 				{
// 					contextLines: []string{"context line 1"},
// 					oldLines:     []string{"removed line 1"},
// 					newLines:     []string{"added line 1"},
// 				},
// 				{
// 					contextLines: []string{"context line 2"},
// 					oldLines:     []string{"removed line 2"},
// 					newLines:     []string{"added line 2"},
// 				},
// 				{
// 					contextLines: []string{"context line 3"},
// 					oldLines:     []string{},
// 					newLines:     []string{},
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := groupRelatedChanges(tt.lines)

// 			assert.Equal(t, len(tt.expected), len(result))

// 			for i, expectedGroup := range tt.expected {
// 				if i < len(result) {
// 					actualGroup := result[i]
// 					assert.Equal(t, expectedGroup.contextLines, actualGroup.contextLines)
// 					assert.Equal(t, expectedGroup.oldLines, actualGroup.oldLines)
// 					assert.Equal(t, expectedGroup.newLines, actualGroup.newLines)
// 				}
// 			}
// 		})
// 	}
// }
