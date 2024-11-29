package graph

import (
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"ralts-cms/graph/model"
	"testing"
)

type ResolverTestSuite struct {
	suite.Suite
}

func (suite *ResolverTestSuite) SetupTest() {}

func (suite *ResolverTestSuite) TearDownTest() {}

func TestQueryResolver_Machines(t *testing.T) {
	c := client.New(handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: &Resolver{}})))
	q := `
      query GetMachines { 
        machines {
          serialNumber
          customer
		  accountType
		  reportedBy
		  ppmDate
        }
      }
    `
	var resp struct {
		Machines []*model.Machine
	}
	c.MustPost(q, &resp)
	require.Len(t, resp.Machines, 1)
	assert.Equal(t, resp.Machines[0].SerialNumber, "123-A")
}

func TestQueryResolver_MachineForm(t *testing.T) {
	c := client.New(handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: &Resolver{}})))
	q := `
      query GetMachineForm { 
        machineForm {
          fields {
			type
		    label
		    value
		  }
        }
      }
    `
	var resp struct {
		MachineForm struct {
			Fields []*model.FormField
		}
	}
	c.MustPost(q, &resp)
	require.Len(t, resp.MachineForm.Fields, 1)
	assert.Equal(t, resp.MachineForm.Fields[0].Type, model.FormFieldTypeText)
	assert.Equal(t, resp.MachineForm.Fields[0].Label, "Serial No.")
	assert.Equal(t, resp.MachineForm.Fields[0].Value, "123-A")
}

func TestResolverTestSuite(t *testing.T) {
	suite.Run(t, new(ResolverTestSuite))
}
