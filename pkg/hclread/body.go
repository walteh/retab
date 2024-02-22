package hclread

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/walteh/terrors"
)

type BodyBuilder struct {
	files map[string]*hclsyntax.Body
}

func (me *BodyBuilder) NewRoot() *hclsyntax.Body {
	root := &hclsyntax.Body{
		Attributes: hclsyntax.Attributes{},
		Blocks:     make([]*hclsyntax.Block, 0),
	}

	for k, v := range me.files {
		sudoblock := &hclsyntax.Block{
			Type:   FilesKey,
			Body:   v,
			Labels: []string{k},
		}
		root.Blocks = append(root.Blocks, sudoblock)
	}

	return root
}

func (me *BodyBuilder) NewRootForFile(file string) (*hclsyntax.Body, error) {

	if me.files[file] == nil {
		return nil, terrors.Errorf("file %s not found", file)
	}

	root := me.NewRoot()
	for k, v := range me.files[file].Attributes {
		root.Attributes[k] = v
	}

	root.Blocks = append(root.Blocks, me.files[file].Blocks...)

	return root, nil
}

func (me *BodyBuilder) GetAllBlocksOfType(name string) []*hclsyntax.Block {
	var out []*hclsyntax.Block

	for _, v := range me.files {
		for _, blk := range v.Blocks {
			if blk.Type == name {
				out = append(out, blk)
			}
		}
	}

	return out
}
