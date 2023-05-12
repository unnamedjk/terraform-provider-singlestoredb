package workspacegroups_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/singlestore-labs/singlestore-go/management"
	"github.com/singlestore-labs/terraform-provider-singlestoredb/examples"
	"github.com/singlestore-labs/terraform-provider-singlestoredb/internal/provider/config"
	"github.com/singlestore-labs/terraform-provider-singlestoredb/internal/provider/testutil"
	"github.com/singlestore-labs/terraform-provider-singlestoredb/internal/provider/util"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestReadsWorkspaceGroup(t *testing.T) {
	workspaceGroup := management.WorkspaceGroup{
		AllowAllTraffic: nil,
		CreatedAt:       "2023-02-28T05:33:06.3003Z",
		ExpiresAt:       nil,
		FirewallRanges:  util.Ptr([]string{"127.0.0.1/32"}),
		Name:            "foo",
		RegionID:        uuid.MustParse("0aa1aff3-4092-4a0c-bf36-da54e85a4fdf"),
		State:           management.ACTIVE,
		TerminatedAt:    nil,
		UpdateWindow: &management.UpdateWindow{
			Day:  3,
			Hour: 15,
		},
		WorkspaceGroupID: uuid.MustParse("e1a0a960-8591-4196-bb26-f53f0f8e35ce"),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, fmt.Sprintf("/v1/workspaceGroups/%s", workspaceGroup.WorkspaceGroupID), r.URL.Path)
		w.Header().Add("Content-Type", "json") // Necessary to make the library parse the resulting JSON.
		_, err := w.Write(testutil.MustJSON(workspaceGroup))
		require.NoError(t, err)
	}))
	defer server.Close()

	testutil.UnitTest(t, testutil.UnitTestConfig{
		APIServiceURL: server.URL,
		APIKey:        testutil.UnusedAPIKey,
	}, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testutil.UpdatableConfig(examples.WorkspaceGroupsGetDataSource).
					WithWorkspaceGroupGetDataSoure("example")(config.IDAttribute, cty.StringVal(workspaceGroup.WorkspaceGroupID.String())).
					String(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", config.IDAttribute, workspaceGroup.WorkspaceGroupID.String()),
					resource.TestCheckNoResourceAttr("data.singlestoredb_workspace_group.example", "allow_all_traffic"),
					resource.TestCheckNoResourceAttr("data.singlestoredb_workspace_group.example", "expires_at"),
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", "firewall_ranges.#",
						strconv.Itoa(len(util.Deref(workspaceGroup.FirewallRanges))),
					),
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", "name", workspaceGroup.Name),
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", "region_id", workspaceGroup.RegionID.String()),
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", "state", string(workspaceGroup.State)),
					resource.TestCheckNoResourceAttr("data.singlestoredb_workspace_group.example", "terminated_at"),
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", "update_window.day",
						strconv.Itoa(int(workspaceGroup.UpdateWindow.Day)),
					),
					resource.TestCheckResourceAttr("data.singlestoredb_workspace_group.example", "update_window.hour",
						strconv.Itoa(int(workspaceGroup.UpdateWindow.Hour)),
					),
				),
			},
		},
	})
}

func TestWorkspaceGroupNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	testutil.UnitTest(t, testutil.UnitTestConfig{
		APIServiceURL: server.URL,
		APIKey:        "bar",
	}, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testutil.UpdatableConfig(examples.WorkspaceGroupsGetDataSource).
					WithWorkspaceGroupGetDataSoure("example")(config.IDAttribute, cty.StringVal(uuid.New().String())).
					String(),
				ExpectError: regexp.MustCompile(http.StatusText(http.StatusNotFound)),
			},
		},
	})
}

func TestInvalidInputUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.False(t, true, "should not get here")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	testutil.UnitTest(t, testutil.UnitTestConfig{
		APIServiceURL: server.URL,
		APIKey:        "bar",
	}, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testutil.UpdatableConfig(examples.WorkspaceGroupsGetDataSource).
					WithWorkspaceGroupGetDataSoure("example")(config.IDAttribute, cty.StringVal("valid-uuid")).
					String(),
				ExpectError: regexp.MustCompile("invalid UUID"),
			},
		},
	})
}

func TestGetWorkspaceGroupNotFoundIntegration(t *testing.T) {
	testutil.IntegrationTest(t, testutil.IntegrationTestConfig{
		APIKey: os.Getenv(config.EnvTestAPIKey),
	}, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testutil.UpdatableConfig(examples.WorkspaceGroupsGetDataSource).
					WithWorkspaceGroupGetDataSoure("example")(config.IDAttribute, cty.StringVal(uuid.New().String())).
					String(),
				ExpectError: regexp.MustCompile(http.StatusText(http.StatusNotFound)), // Checking that at least the expected error.
			},
		},
	})
}
